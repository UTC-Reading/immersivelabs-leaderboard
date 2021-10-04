package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/term"
)

func main() {
	c := &http.Client{
		Timeout: time.Second * 10,
	}

	account := Account{}

	if err := account.collectCredentials(); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("\n\033[33mSigning In...\033[0m")
	if err := account.login(c); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("\n\033[33mFetching data...\033[0m")
	ldbRes, err := account.getLeaderboard(c, "null")
	if err != nil {
		log.Fatalln(err)
	}

	totalUsers := strconv.Itoa(ldbRes.Data.LdbConn.TotalCount)
	fmt.Println("There are " + totalUsers + " users in your organisation")

	var numTopUsers string
	fmt.Print("Number of users (default=all, highest to lowest): ")
	fmt.Scanln(&numTopUsers)

	switch numTopUsers {
	case "all":
	case "":
		ldbRes, _ = account.getLeaderboard(c, totalUsers)
	default:
		if _, err := strconv.Atoi(numTopUsers); err != nil {
			log.Fatalln(err)
		}
		ldbRes, _ = account.getLeaderboard(c, numTopUsers)
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	downloaddir := homedir + "/Downloads/ImmersiveLabs"
	// If download directory doesn't exist, create it
	if _, err := os.Stat(downloaddir); os.IsNotExist(err) {
		os.MkdirAll(downloaddir, 0700)
	}

	filename := downloaddir + "/Leaderboard" + time.Now().String() + ".csv"
	if err := ldbRes.writeToCSVFile(filename); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("\033[32m", "File saved to \""+filename+"\" succesfully", "\033[0m")
	fmt.Scanln("Press any key to exit")
}

type Account struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"-"`
}

func (account *Account) collectCredentials() error {

	fmt.Print("Email Address: ")
	fmt.Scanln(&account.Email)

	fmt.Print("Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}

	account.Password = string(bytePassword)

	return nil
}

func (account *Account) login(c *http.Client) error {
	authURL := "https://api.immersivelabs.online/v1/immersive_auth/sessions"

	requestData, _ := json.Marshal(struct {
		Account Account `json:"account"`
	}{
		Account: *account,
	})

	req, err := http.NewRequest("POST", authURL, bytes.NewReader(requestData))
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Des", "empty")
	req.Header.Add("Sec-GPC", "1")
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}

	response, err := c.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	token := response.Header.Get("Authorization")
	if token == "" {
		return errors.New("Incorrect Credentials")
	}

	account.Token = token
	return nil
}

func (account *Account) getLeaderboard(c *http.Client, numTopUsers string) (*LeaderboardResponse, error) {
	leaderboardURL := "https://api.immersivelabs.online/v1/graphql"

	requestData := fmt.Sprintf(`{"variables":{"first":%s,"leaderboardContext":{"type":"ORGANISATIONAL"},"participantType":"ACCOUNT"},"query":"query GetLeaderboardTableData(  $leaderboardContext: LeaderboardContextInput  $participantType: LeaderboardParticipant  $limit: Int  $after: String = null  $before: String = null  $first: Int = null  $last: Int = null) {  ...GetLeaderboardData}fragment GetLeaderboardData on Query {  leaderboardConnection(    profileId: null    leaderboardContext: $leaderboardContext    participantType: $participantType    landOnParticipantPage: false    limit: $limit    after: $after    before: $before    first: $first    last: $last  ) {    totalCount    edges {      position      node {        id        title        points        profileAvatar {          versions {            w100h100          }        }      }    }  }}"}`, numTopUsers)

	req, err := http.NewRequest("POST", leaderboardURL, bytes.NewReader([]byte(requestData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Des", "empty")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", account.Token)

	response, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	res, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	result := &LeaderboardResponse{}
	json.Unmarshal(res, result)

	return result, nil
}

type User struct {
	Position int
	Username string
	Points   int
}

func (user User) GetHeaders() []string {
	return []string{"Position", "Username", "Points"}
}

func (user User) ToSlice() []string {
	return []string{strconv.Itoa(user.Position), user.Username, strconv.Itoa(user.Points)}
}

type LeaderboardResponse struct {
	Data struct {
		LdbConn struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Position int `json:"position"`
				Node     struct {
					Username string `json:"title"`
					Points   int    `json:"points"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"leaderboardConnection"`
	} `json:"data"`
}

func (ldbRes *LeaderboardResponse) writeToCSVFile(filepath string) error {
	users := ldbRes.getUsers()
	records := [][]string{}
	headers := User{}.GetHeaders()
	records = append(records, headers)

	for _, user := range users {
		records = append(records, user.ToSlice())
	}

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	w := csv.NewWriter(f)
	w.WriteAll(records) // calls Flush internally

	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

func (ldbRes *LeaderboardResponse) getUsers() []User {
	var users []User
	for _, value := range ldbRes.Data.LdbConn.Edges {
		user := User{
			Position: value.Position,
			Username: value.Node.Username,
			Points:   value.Node.Points,
		}

		users = append(users, user)
	}
	return users
}
