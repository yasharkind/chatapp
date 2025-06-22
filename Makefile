build: pull
	@echo building...
	@go build -mod=vendor && echo success

pull:
	@echo pulling master...
	@git pull origin master
	@echo
