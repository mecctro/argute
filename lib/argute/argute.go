package argute

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Commands : All bot commands
type Commands []struct {
	Command string
	//Parameters [...]string
}

// Users : All users
type Users []struct {
	Mode string
	Name string
	Rank string
}

// Insult : remote insult
type Insult struct {
	Insult    string `json:"insult"`
	Source    string `json:"source"`
	SourceURL string `json:"sourceUrl"`
}

// Chuck : remote chuck norris joke
type Chuck struct {
	Type  string `json:"type"`
	Value struct {
		ID         int           `json:"id"`
		Joke       string        `json:"joke"`
		Categories []interface{} `json:"categories"`
	} `json:"value"`
}

// Quote : remote memorable quote
type Quote struct {
	QuoteText   string `json:"quoteText"`
	QuoteAuthor string `json:"quoteAuthor"`
	SenderName  string `json:"senderName"`
	SenderLink  string `json:"senderLink"`
	QuoteLink   string `json:"quoteLink"`
}

// Cookie : whole remote fortune cookie
type Cookie []struct {
	Fortune struct {
		Message string `json:"message"`
		ID      string `json:"id"`
	} `json:"fortune"`
	Lesson struct {
		English       string `json:"english"`
		Chinese       string `json:"chinese"`
		Pronunciation string `json:"pronunciation"`
		ID            string `json:"id"`
	} `json:"lesson"`
	Lotto struct {
		ID      string `json:"id"`
		Numbers []int  `json:"numbers"`
	} `json:"lotto"`
}

var (
	// AllCommands : Stores all bot commands
	AllCommands = Commands{}

	// AllUsers : Stores all users
	AllUsers = Users{}
)

func init() {

}

// GetInsult : Gets an insult
func GetInsult() Insult {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://quandyfactory.com/insult/json", nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	req.Header.Add("Accept", "application/json")

	decoder := json.NewDecoder(resp.Body)
	v := Insult{}
	err = decoder.Decode(&v)
	if err != nil {
		fmt.Println(err)
	}

	return v
}

// GetChuck : Gets a chuck norris joke
func GetChuck() Chuck {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://api.icndb.com/jokes/random", nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	req.Header.Add("Accept", "application/json")

	decoder := json.NewDecoder(resp.Body)
	v := Chuck{}
	err = decoder.Decode(&v)
	if err != nil {
		fmt.Println(err)
	}

	return v
}

// GetQuote : Gets a memorable quote
func GetQuote() Quote {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://api.forismatic.com/api/1.0/?method=getQuote&format=json&lang=en", nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	req.Header.Add("Accept", "application/json")

	decoder := json.NewDecoder(resp.Body)
	v := Quote{}
	err = decoder.Decode(&v)
	if err != nil {
		fmt.Println(err)
	}

	return v
}

// GetCookie : Gets a fortune cookie
func GetCookie() Cookie {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://fortunecookieapi.com/v1/cookie", nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	req.Header.Add("Accept", "application/json")

	decoder := json.NewDecoder(resp.Body)
	v := Cookie{}
	err = decoder.Decode(&v)
	if err != nil {
		fmt.Println(err)
	}

	return v
}
