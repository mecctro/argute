package fah

import (
	"compress/bzip2"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"time"
)

// Team : A group of users
type Team struct {
	ID    int
	Name  string
	Score int
	WU    int
}

// Teams : Group of teams.
type Teams struct {
	Teams []Team
}

// User : A single user
type User struct {
	Name   string
	Score  int
	WU     int
	TeamID int
}

// Users : Group of users.
type Users struct {
	Users []User
}

type queueInfo []struct {
	ID             string    `json:"id"`
	State          string    `json:"state"`
	Error          string    `json:"error"`
	Project        int       `json:"project"`
	Run            int       `json:"run"`
	Clone          int       `json:"clone"`
	Gen            int       `json:"gen"`
	Core           string    `json:"core"`
	Unit           string    `json:"unit"`
	Percentdone    string    `json:"percentdone"`
	Eta            string    `json:"eta"`
	Ppd            string    `json:"ppd"`
	Creditestimate string    `json:"creditestimate"`
	Waitingon      string    `json:"waitingon"`
	Nextattempt    string    `json:"nextattempt"`
	Timeremaining  string    `json:"timeremaining"`
	Totalframes    int       `json:"totalframes"`
	Framesdone     int       `json:"framesdone"`
	Assigned       time.Time `json:"assigned"`
	Timeout        time.Time `json:"timeout"`
	Deadline       time.Time `json:"deadline"`
	Ws             string    `json:"ws"`
	Cs             string    `json:"cs"`
	Attempts       int       `json:"attempts"`
	Slot           string    `json:"slot"`
	Tpf            string    `json:"tpf"`
	Basecredit     string    `json:"basecredit"`
}

const (
	userSummaryFile = "daily_user_summary.txt.bz2"
	teamSummaryFile = "daily_team_summary.txt.bz2"

	userSummaryPath = "./fah/stats/" + userSummaryFile
	teamSummaryPath = "./fah/stats/" + teamSummaryFile

	userUpdatePath = "http://fah-web.stanford.edu/" + userSummaryFile
	teamUpdatePath = "http://fah-web.stanford.edu/" + teamSummaryFile
)

var (
	// AllTeams : All F@H teams.
	AllTeams = Teams{}
	// AllUsers : All F@H users.
	AllUsers = Users{}
)

// Team sorting
func (fah Teams) Len() int {
	return len(fah.Teams)
}

func (fah Teams) Less(i, j int) bool {
	return fah.Teams[i].Score > fah.Teams[j].Score
}

func (fah Teams) Swap(i, j int) {
	fah.Teams[i], fah.Teams[j] = fah.Teams[j], fah.Teams[i]
}

// User sorting
func (fah Users) Len() int {
	return len(fah.Users)
}

func (fah Users) Less(i, j int) bool {
	return fah.Users[i].Score > fah.Users[j].Score
}

func (fah Users) Swap(i, j int) {
	fah.Users[i], fah.Users[j] = fah.Users[j], fah.Users[i]
}

// AddTeam : Append a team.
func (fah *Teams) AddTeam(thisTeam Team) Team {
	fah.Teams = append(fah.Teams, thisTeam)
	return thisTeam
}

// AddUser : Append a user.
func (fah *Users) AddUser(thisUser User) User {
	fah.Users = append(fah.Users, thisUser)
	return thisUser
}

// GetTeamByName : Get a specific teams data by name.
func (fah *Teams) GetTeamByName(name string) (Team, int) {
	for i, thisTeam := range fah.Teams {
		if thisTeam.Name == name {
			return thisTeam, i
		}
	}

	return Team{0, "nothing", 0, 0}, 0
}

// GetUserByName : Get a specific users data by name.
func (fah *Users) GetUserByName(name string) (User, int) {
	for i, thisUser := range fah.Users {
		if thisUser.Name == name {
			return thisUser, i
		}
	}

	return User{"nothing", 0, 0, 0}, 0
}

// GetUsersByTeamTopRank : Get a users data by team rank.
func (fah *Users) GetUsersByTeamTopRank(teamID int) []User {
	tempUsers := Users{}
	i := 0
	for _, thisUser := range fah.Users {
		if thisUser.TeamID == teamID {
			if i < 10 {
				tempUsers.AddUser(thisUser)
				i++
			}
		}
	}

	sort.Sort(tempUsers)
	return tempUsers.Users
}

// LoadTeamSummary : Loads the team summary flat-file.
func LoadTeamSummary() {
	teamSummaryTxt := readSummary(teamSummaryPath)

	// Split all teams in summary.
	var r = regexp.MustCompile("([0-9]*)\t([a-zA-Z`~!@#$%^&*\\\\()_+\\\\-|}\\]{\\[\"':;?\\/>.<,\\s]*)\t([0-9]*)\t([0-9]*)")
	res := r.FindAllStringSubmatch(teamSummaryTxt, -1)

	// Handle and assign team data
	for i := range res {

		ID, _ := strconv.Atoi(res[i][1])
		Name := res[i][2]
		Score, _ := strconv.Atoi(res[i][3])
		WU, _ := strconv.Atoi(res[i][4])

		AllTeams.AddTeam(Team{
			ID,
			Name,
			Score,
			WU,
		})
	}
	sort.Sort(AllTeams)
	debug.FreeOSMemory()
}

// LoadUserSummary : Loads the User summary flat-file.
func LoadUserSummary() {
	userSummaryTxt := readSummary(userSummaryPath)

	// Split all users in summary.
	var r = regexp.MustCompile("([0-9a-zA-Z`~!@#$%^&*\\\\()_+\\-|}\\]{\\[\"':;?\\/>.<,\\s]*?)\\t([0-9]*)\\t([0-9]*)\\t([0-9]*)\\n")
	res := r.FindAllStringSubmatch(userSummaryTxt, -1)

	// Handle and assign user data
	for i := range res {

		Name := res[i][1]
		Score, _ := strconv.Atoi(res[i][2])
		WU, _ := strconv.Atoi(res[i][3])
		TeamID, _ := strconv.Atoi(res[i][4])

		AllUsers.AddUser(User{
			Name,
			Score,
			WU,
			TeamID,
		})
	}
	sort.Sort(AllUsers)
	debug.FreeOSMemory()
}

// UpdateTeamSummary : Downloads and upacks team data.
func UpdateTeamSummary() {
	//if _, err := os.Stat(teamSummaryBzip); os.IsNotExist(err) {
	in, err := os.Create(teamSummaryPath)
	if err != nil {
		fmt.Println(err)
	}
	defer in.Close()

	// Get the data
	resp, err := http.Get(teamUpdatePath)
	if err != nil {
		fmt.Println(err)
	} else {
		defer resp.Body.Close()

		// Writer the body to file
		_, err = io.Copy(in, resp.Body)
		if err != nil {
			fmt.Println(err)
		}
	}

	AllTeams = Teams{}
	LoadTeamSummary()
}

// UpdateUserSummary : Downloads and upacks user data.
func UpdateUserSummary() {
	//if _, err := os.Stat(userSummaryBzip); os.IsNotExist(err) {
	in, err := os.Create(userSummaryPath)
	if err != nil {
		fmt.Println(err)
	}
	defer in.Close()

	// Get the data
	resp, err := http.Get(userUpdatePath)
	if err != nil {
		fmt.Println(err)
	} else {
		defer resp.Body.Close()

		// Writer the body to file
		_, err = io.Copy(in, resp.Body)
		if err != nil {
			fmt.Println(err)
		}
	}

	AllUsers = Users{}
	LoadUserSummary()
}

func readSummary(summaryBzip string) string {
	// Temp storage
	summaryText := ""
	in, err := os.Open(summaryBzip)
	if err != nil {
		fmt.Println(err)
	} else {
		defer in.Close()

		bzip := bzip2.NewReader(in)
		buf := make([]byte, 0, 4*1024)
		for {
			n, err := bzip.Read(buf[:cap(buf)])
			buf = buf[:n]
			if n == 0 {
				if err == nil {
					continue
				}
				if err == io.EOF {
					break
				}
			}
			// process buf
			if err != nil && err != io.EOF {
				fmt.Println(err)
			} else {
				summaryText += string(buf)
			}
		}
	}
	return summaryText
}
