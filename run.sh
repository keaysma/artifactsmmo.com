# Evaluate .env file into shell env (temporarily, for just this line)
# Run hello.go
env $(grep -v "^ *#" -E .env | xargs) go run hello.go