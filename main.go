package main

var (
	Environment Env
	Config      Conf
)

func init() {
	Environment = initEnvironment()
	initConfig()
}

func main() {

}
