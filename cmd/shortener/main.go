package main

func main() {
	conf := WebConfig{}
	conf.ParseParams()

	listenAndServe(conf)
}
