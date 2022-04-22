package main

func main() {
	// fmt.Println("启动")
	server := NewServer("127.0.0.1", 8888)
	server.Start()
}
