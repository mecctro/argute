package argute

// Commands : All bot commands
type Commands []struct {
	Command    string
	Parameters [...]string
}

// Users : All users
type Users []struct {
	Mode string
	Name string
	Rank string
}

var (
	// AllCommands : Stores all bot commands
	AllCommands = Commands{}

	// AllUsers : Stores all users
	AllUsers = Users{}
)

func init() {

}
