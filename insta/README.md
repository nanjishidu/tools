# insta

instagram's tool by golang

## Install insta
	
	cd $GOPATH/src
	git clone https://github.com/nanjishidu/tools.git
	cd  tools/insta
	go build insta.go
## Cross Compiling

	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o insta_linux_amd64 insta.go 
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o insta_linux_386 insta.go 
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o insta_linux_arm insta.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o insta_win_amd64 insta.go 
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o insta_win_386 insta.go 
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o insta_mac_amd64 insta.go
## Use insta

	./insta -h

> 	Useage: ./insta [ Options ]
>
>  Options:
> 
>  		-username      instagram username
>  		-imgurl        download the image url
>  		-dir           instagram save path  [Default: uploadir]
>		-http_proxy    http proxy url  [Example: http://127.0.0.1:1087]
>		 -tcp_proxy    tcp proxy  [Example: 127.0.0.1:1086]
>  		-pagesize      number each page to download [Default: 100]
>  		-imgnum        download the image max number
>  		-thumbnails    whether to download a thumbnail
>  		-h             help
>  		-version       insta version




