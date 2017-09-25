default: build

build:
	go build -a -o swan-executor executor.go # TODO strip & compress to make the binary smaller
	mv -fv swan-executor /var/www/html/tmp
