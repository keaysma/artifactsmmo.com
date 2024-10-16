# Evaluate .env file into shell env (temporarily, for just this line)
# Run <first arg>
# Ex: ./run.sh cmd.go
env $(grep -v "^ *#" -E .env | xargs) go run $1 $2