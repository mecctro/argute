package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"./lib/argute"
	"./lib/fah"

	irc "github.com/thoj/go-ircevent"
	telnet "github.com/ziutek/telnet"
)

const (
	server = "irc.soylentnews.org:6697"
	prefix = "`"
	admin  = "mecctro"
)

var (
	channels   = [...]string{"#mecctro" /*"#test", "#folding"*/}
	fahClients = [...]string{"192.168.1.251:36330", "127.0.0.1:36330"}
	fahSlots   = slotInfo{}

	nick               = "argute"
	ircobj             = irc.IRC(nick, nick) //Create new ircobj
	defaultPrivMessage = "You have no power here."
	globalCommandTime  = time.Now()
	globalCommandTimer = 1 * time.Second
	messageCount       = 0

	rnnTrainStatus = false
	rnnOutput      = ""

	chittyStatus = false
	chittyTicker = time.NewTicker(5 * time.Minute)
	chittyQuit   = make(chan struct{})
)

type fahClient struct {
	Name    string
	Address string
	Port    string
}

// F@H remote client machine slots.
type slotInfo []struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Options     struct {
	} `json:"options"`
	Reason string `json:"reason"`
	Idle   bool   `json:"idle"`
}

func main() {
	// Start RNN training
	genCharRNN()
	runCharRNN()
	rnnTicker := time.NewTicker(30 * time.Second)
	rnnQuit := make(chan struct{})
	go func() {
		for {
			select {
			case <-rnnTicker.C:
				if rnnTrainStatus == false {
					go genCharRNN()
				}
				go runCharRNN()
			case <-rnnQuit:
				rnnTicker.Stop()
				return
			}
		}
	}()

	// Load all statistics.
	go fah.LoadTeamSummary()
	go fah.LoadUserSummary()

	// Periodically update local team summary.
	// Update every 1 hour.
	ticker := time.NewTicker(1 * time.Hour)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				go fah.UpdateTeamSummary()
				go fah.UpdateUserSummary()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	//Set options
	//ircobj.VerboseCallbackHandler = true
	//ircobj.Debug = true
	ircobj.UseTLS = true                                     //default is false
	ircobj.TLSConfig = &tls.Config{InsecureSkipVerify: true} //set ssl options
	//ircobj.Password = "[server password]"

	// Connection success callback.
	ircobj.AddCallback("001", func(e *irc.Event) {
		for _, channel := range channels {
			ircobj.Join(channel)
		}
	})

	// Channel join success callback.
	ircobj.AddCallback("366", func(e *irc.Event) {
		//ircobj.Privmsg("#mecctro", "Test Message from SSL\n")
		//ircobj.Quit()
	})

	/*ircobj.AddCallback("366", func(e *irc.Event) {
		dict := "abcdefghijklmnopqrstuvwxyz"

		b := make([]byte, 5)
		for i := range b {
			b[i] = dict[rand.Intn(len(dict))]
		}

		ircobj.Nick(nick + "_" + string(b))
	})*/

	// Commands
	ircobj.AddCallback("PRIVMSG", func(event *irc.Event) {
		go func(event *irc.Event) {
			//fmt.Println(rnnOutput)
			//event.Message() contains the message
			//event.Nick Contains the sender
			//event.Arguments[0] Contains the channel

			/*rank := ""
			if event.Nick == admin {
				rank = "admin"
			}

			if strings.HasPrefix(event.Message(), prefix) {
				// Split on proceeding spaces
				r, _ := regexp.Compile("(.*?)\\s")
				res := r.FindAllStringSubmatch(event.Message(), -1)

				// Check for params and command
				if len(res) > 1 {
					command(rank, event.Arguments[0], res[0][1], res[1:])
				} else {
					// A top level command
					r, _ := regexp.Compile(prefix + "(.*)")
					res := r.FindAllStringSubmatch(event.Message(), -1)

					command(rank, event.Arguments[0], event.Message(), res)
				}
			}*/

			//inputTime := time.Now()
			inputMessage := event.Message()
			outputMessage := ""

			//if event.Arguments[0] == "#mecctro" {
			if inputMessage == "`chitty start" && chittyStatus == false {
				if event.Nick == admin {
					// Cleanup parts, we only want full sentences.
					var r = regexp.MustCompile(".*?\\.\\s(.*[.”\"])\\s")
					res := r.FindAllStringSubmatch(rnnOutput, -1)

					if len(res) > 0 {
						outputMessage = res[0][1]
					} else {
						outputMessage = "sorry!"
					}
					go func() {
						for {
							select {
							case <-chittyTicker.C:
								genCharRNN()

								// Cleanup parts, we only want full sentences.
								var r = regexp.MustCompile(".*?\\.\\s(.*[.”\"])\\s")
								res := r.FindAllStringSubmatch(rnnOutput, -1)

								if len(res) > 0 {
									ircobj.Privmsg(event.Arguments[0], res[0][1])
								} else {
									ircobj.Privmsg(event.Arguments[0], "sorry!")
								}
							case <-chittyQuit:
								chittyTicker.Stop()
								chittyStatus = false
								return
							}
						}
					}()
					chittyStatus = true
				} else {
					outputMessage = "Thou hast no power here."
				}
				//fmt.Println(rnnOutput)
			} else if inputMessage == "`chitty stop" && chittyStatus {
				if event.Nick == admin {
					chittyTicker.Stop()
					chittyStatus = false
					outputMessage = "Chitty stopped.\n"
				} else {
					outputMessage = "Thou hast no power here."
				}
			} else if inputMessage == "`chitty gen" {
				if event.Nick == admin {
					genCharRNN()

					// Cleanup parts, we only want full sentences.
					/*var r = regexp.MustCompile(".*?\\.\\s(.*[.”\"])\\s")
					res := r.FindAllStringSubmatch(rnnOutput, -1)

					if len(res) > 0 {
						outputMessage = res[0][1]
					} else {
						outputMessage = "sorry!"
					}*/
					outputMessage = rnnOutput
				} else {
					outputMessage = "Thou hast no power here."
				}
			}
			//}

			//if inputTime.After(globalCommandTime.Add(globalCommandTimer)) {
			switch {

			case strings.Compare(inputMessage, prefix+"help") == 0:
				outputMessage = "Commands: `help (returns this) | `quote (returns an interesting quote) | `chuck (returns Chuck Norris joke) | `cookie (returns a fortune cookie) | `fah <stat> <user> (returns stat of user; eg. rank, score, work ) | `fah team (returns our F@H team stats)\n"

			case strings.HasPrefix(inputMessage, prefix+"join"):
				// Check privelage
				if event.Nick == admin {
					r, _ := regexp.Compile("(#.*?)\\s")
					res := r.FindAllStringSubmatch(event.Message(), -1)

					// Check for parameters.
					if len(res) > 0 {
						outputMessage = "Joined: "
						for _, channel := range res {
							ircobj.Join(channel[1])
							outputMessage += channel[1] + " "
						}
					}
				} else {
					outputMessage = defaultPrivMessage
				}

			case strings.HasPrefix(inputMessage, prefix+"part"):
				// Check privelage
				if event.Nick == admin {
					r, _ := regexp.Compile("(#.*?)\\s")
					res := r.FindAllStringSubmatch(event.Message(), -1)

					// Check for parameters.
					if len(res) > 0 {
						outputMessage = "Parted: "
						for _, channel := range res {
							ircobj.Part(channel[1])
							outputMessage += channel[1] + " "
						}
					} else {
						ircobj.Part(event.Arguments[0])
					}
				} else {
					outputMessage = defaultPrivMessage
				}

			case strings.HasPrefix(inputMessage, prefix+"insult"):
				if event.Nick == admin {
					r, _ := regexp.Compile(prefix + "insult\\s(.*)")
					res := r.FindAllStringSubmatch(inputMessage, -1)

					v := argute.GetInsult()
					if v.Insult != "" && len(res) > 0 {
						outputMessage = res[0][1] + ": " + v.Insult
					}
				} else {
					outputMessage = "Thou hast no power here."
				}

			case strings.Compare(inputMessage, prefix+"chuck") == 0:
				v := argute.GetChuck()
				if v.Value.Joke != "" {
					outputMessage = strings.Replace(v.Value.Joke, "&quot;", "\"", -1)
				} else {
					outputMessage = "Sorry, I couldn't get that right now."
				}

			case strings.Compare(inputMessage, prefix+"quote") == 0:

				v := argute.GetQuote()
				if v.QuoteText != "" {
					outputMessage = "\"" + v.QuoteText + "\" : " + v.QuoteAuthor
				}

			case strings.Compare(inputMessage, prefix+"cookie") == 0:
				v := argute.GetCookie()
				if v[0].Fortune.Message != "" {
					outputMessage = "Fortune: " + v[0].Fortune.Message + " | Learn Chinese: " + v[0].Lesson.English + " / " + v[0].Lesson.Chinese + " (" + v[0].Lesson.Pronunciation + ") | "
					lotto := ""
					for _, number := range v[0].Lotto.Numbers {
						lotto += " " + strconv.Itoa(number)
					}
					outputMessage += "Lotto:" + lotto
				}

			case strings.Compare(inputMessage, prefix+"fah links") == 0:
				outputMessage = "Official SoylentNews.org Folding Team http://fah-web.stanford.edu/cgi-bin/main.py?qtype=teampage&teamnum=230319 | Get started here: http://folding.stanford.edu | Better Stats at http://folding.extremeoverclocking.com/team_summary.php?s=&t=230319"

			case strings.Compare(inputMessage, prefix+"fah update") == 0:
				// Check privelage
				if event.Nick == admin {
					ircobj.Privmsg(event.Arguments[0], "Updating F@H Stats...")
					fah.UpdateTeamSummary()
					fah.UpdateUserSummary()
					ircobj.Privmsg(event.Arguments[0], "Finished.")
				} else {
					outputMessage = defaultPrivMessage
				}

			case strings.Compare(inputMessage, prefix+"fah team top") == 0:
				// Pull teams specific daily statistics
				team, _ := fah.AllTeams.GetTeamByName("SoylentNews.org")
				users := fah.AllUsers.GetUsersByTeamTopRank(team.ID)

				if len(users) > 0 {
					outputMessage = "FAH Top 10 for " + team.Name + ": "
					for _, user := range users {
						outputMessage += user.Name + ", "
					}
				}

				//outputMessage = "Top contributers for: " + team.Name + ", " + strconv.Itoa() + ") | Rank: " + strconv.Itoa(teamRank) + " of " + strconv.Itoa(fah.AllTeams.Len()) + ", Score: " + strconv.Itoa(team.Score) + ", Work Units: " + strconv.Itoa(team.WU) + "\n"

			case strings.Compare(inputMessage, prefix+"fah team") == 0:
				// Pull teams specific daily statistics
				team, teamRank := fah.AllTeams.GetTeamByName("SoylentNews.org")
				outputMessage = "Stats for: " + team.Name + " (" + strconv.Itoa(team.ID) + ") | Rank: " + strconv.Itoa(teamRank) + " of " + strconv.Itoa(fah.AllTeams.Len()) + ", Score: " + strconv.Itoa(team.Score) + ", Work Units: " + strconv.Itoa(team.WU) + "\n"

			case strings.Compare(inputMessage, prefix+"fah slots") == 0:
				// Check privelage
				if event.Nick == admin {
					slots := ""
					fahSlots = slotInfo{}
					for _, client := range fahClients {
						getFAHClientSlots(client)
					}
					for _, slot := range fahSlots {
						slots += " " + slot.Description + " - " + slot.Status + ","
					}
					outputMessage = "Slots:" + slots + "\n"
				} else {
					outputMessage = defaultPrivMessage
				}

			case strings.HasPrefix(inputMessage, prefix+"fah"):
				// Folding @ Home commands

				// Search for and return fah commands
				r, _ := regexp.Compile(prefix + "fah\\s(.*)\\s(.*)")
				res := r.FindAllStringSubmatch(event.Message(), -1)

				// Check for both parameters.
				if len(res) > 0 {
					command := res[0][1]
					username := res[0][2]

					// User Details.
					user, userRank := fah.AllUsers.GetUserByName(username)
					totalUsers := fah.AllUsers.Len()
					userRankPercent := (float64(userRank) / float64(totalUsers)) * 100

					// Make sure user doesn't contain illegal characters.
					r, _ = regexp.Compile("[^\\w]")
					if r.MatchString(user.Name) || len(user.Name) > 50 {
						outputMessage = "That isn't a valid fah username!"
					} else {
						// Check user has done work.
						if user.Name == "nothing" {
							outputMessage = user.Name + " doesn't exist, or hasn't done any work!\n"
						} else {
							// Each fah command.
							switch command {
							case "score":
								outputMessage = "Score for " + user.Name + ": " + strconv.Itoa(user.Score) + "\n"
							case "rank":
								outputMessage = "Rank for " + user.Name + ": " + strconv.Itoa(userRank) + " of " + strconv.Itoa(totalUsers) + " (top " + strconv.FormatFloat(userRankPercent, 'f', 2, 64) + "%)\n"
							case "work":
								outputMessage = "Work Units for " + user.Name + ": " + strconv.Itoa(user.WU) + "\n"
							default:
								break
							}
						}
					}
				}

			}

			if outputMessage != "" {
				ircobj.Privmsg(event.Arguments[0], outputMessage)
			} else {
				saveCharRNNLog(event.Message())
				//ircobj.Privmsg(event.Arguments[0], "Sorry, I couldn't get that right now.")
			}
			//} else {
			//	globalCommandTime = time.Now()
			//}
		}(event)
	})

	err := ircobj.Connect(server)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		/*err = */ ircobj.Loop()
		/*if err != nil {
			ircobj = irc.IRC(nick, nick)
			err = ircobj.Connect(server)
			if err != nil {
				fmt.Println(err.Error())
			}
		}*/
	}
}

func saveCharRNNLog(input string) {
	in, err := os.OpenFile("./lib/bin/input/irc.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer in.Close()

	// Write the string to file
	_, err = io.WriteString(in, input+" ")
	if err != nil {
		fmt.Println(err)
	}
	//}
}

func runCharRNN() {
	if rnnTrainStatus == false {
		execute := "./lib/bin/char-rnn.exe"
		args := []string{"train", "lstm", "./lib/bin/output/lstm", "./lib/bin/input/"}
		env := os.Environ()
		//runtime.GOMAXPROCS(2)
		env = append(env, "GOMAXPROCS=3")
		//env := []string("GOMAXPROCS=2")
		cmd := exec.Command(execute, args...)
		cmd.Env = env
		//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

		stdin, err := cmd.StdinPipe()
		if err != nil {
			fmt.Println(err)
		}

		/*stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Println(err)
		}*/

		stderr, err := cmd.StderrPipe()
		if err != nil {
			fmt.Println(err)
		}

		err = cmd.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
			return
		}

		// Don't let main() exit before our command has finished running
		//defer cmd.Wait() // Doesn't block

		cmderr := bufio.NewScanner(stderr)
		go func() {
			for cmderr.Scan() {
				if strings.Contains(cmderr.Text(), "Training") && rnnTrainStatus == false {
					rnnTrainStatus = true
				}
				if strings.Contains(cmderr.Text(), "Epoch 0") {
					stdin.Write([]byte("\nexit\n"))
				}
				if strings.Contains(cmderr.Text(), "Exit") {
					rnnTrainStatus = false
				}
				fmt.Println(cmderr.Text())
			}
			if err := cmderr.Err(); err != nil {
				fmt.Println(err)
			}
		}()
	}
}

func genCharRNN() {
	execute := "./lib/bin/char-rnn.exe"
	args := []string{"gen", "./lib/bin/output/lstm", "250"}
	env := os.Environ()
	//runtime.GOMAXPROCS(2)
	env = append(env, "GOMAXPROCS=2")
	cmd := exec.Command(execute, args...)
	cmd.Env = env
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		return
	}

	// Don't let main() exit before our command has finished running
	//defer cmd.Wait() // Doesn't block

	cmderr := bufio.NewScanner(stdout)
	go func() {
		rnnOutput = ""
		for cmderr.Scan() {
			rnnOutput += cmderr.Text()
			//fmt.Println(cmderr.Text())
		}
		if err := cmderr.Err(); err != nil {
			fmt.Println(err)
		}
	}()
}

func getFAHClientSlots(address string) {
	data := ""
	client, err := telnet.Dial("tcp", address)
	if err != nil {
		fmt.Println(err)
	} else {

		client.SetUnixWriteMode(true)

		client.Write([]byte("auth x\n"))
		client.Write([]byte("slot-info\n"))

		for {
			read, err2 := client.ReadString('\n')
			if err2 != nil || err2 == io.EOF {
				fmt.Println(err2)
				break
			} else {
				if read == "---\n" {
					break
				} else {
					data += read
				}
			}
		}

		client.Close()

		var r = regexp.MustCompile("(?s)(\\[[^\\w].*\\])")
		res := r.FindAllStringSubmatch(data, -1)

		// Fix incoming json.
		replace := strings.NewReplacer("True", "true", "False", "false")
		data = replace.Replace(res[0][1])

		slotJSON := slotInfo{}
		err = json.Unmarshal([]byte(data), &slotJSON)
		if err != nil {
			fmt.Println(err)
		}

		for _, slot := range slotJSON {
			fahSlots = append(fahSlots, slot)
		}
	}
}
