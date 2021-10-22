package buildinfo

const (
	Graffiti       = "       .__                     __  .__                        ___.           __   \n  ____ |  |__   _____ ___.__._/  |_|__| _____   ____          \\_ |__   _____/  |_ \n /  _ \\|  |  \\ /     <   |  |\\   __\\  |/     \\_/ __ \\   ______ | __ \\ /  _ \\   __\\\n(  <_> )   Y  \\  Y Y  \\___  | |  | |  |  Y Y  \\  ___/  /_____/ | \\_\\ (  <_> )  |  \n \\____/|___|  /__|_|  / ____| |__| |__|__|_|  /\\___  >         |___  /\\____/|__|  \n            \\/      \\/\\/                    \\/     \\/              \\/             "
	GreetingCLI    = "\nversion: %s \nbuild time: %s\ntg: %s\ngithub: %s\n"
	GithubBloopURL = "https://github.com/robotomize/ohmytime-bot.git"
	TgBloopURL     = "https://t.me/ohmytimebot"
)

var (
	BuildTag = "v0.0.0"
	Name     = "ohmytime-bot"
	Time     = ""
)

type buildinfo struct{}

func (buildinfo) Tag() string {
	return BuildTag
}

func (buildinfo) Name() string {
	return Name
}

func (buildinfo) Time() string {
	return Time
}

var Info buildinfo
