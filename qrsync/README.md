# qrsync

qiniu sync tool by golang

## Install qrsync
	
	cd $GOPATH/src
	git clone https://github.com/nanjishidu/tools.git
	go get -u qiniupkg.com/api.v7
	cd  tools/qrsync
	go build qrsync.go
## Cross Compiling

	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o qrsync_linux_amd64 qrsync.go 
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o qrsync_linux_386 qrsync.go 
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o qrsync_linux_arm qrsync.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o qrsync_win_amd64 qrsync.go 
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o qrsync_win_386 qrsync.go 
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o qrsync_mac_amd64 qrsync.go
## Use qrsync

	./qrsync -h

> 	Useage: ./qrsync [ Options ]
>
>  Options:
> 
>	 	-dir           dir [Default: ./]
>  		-ak            qiniu ak (require)
> 		 -sk            qiniu sk (require)
>  		-bucket        qiniu bucket (require)[Example: bucket1|bucket2]
>  		-h             help
>  		-version       qrsync version




