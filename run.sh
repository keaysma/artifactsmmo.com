# Evaluate .env file into shell env (temporarily, for the execution of this script)
cat .env | while read line; do
    if ! [[ $line =~ "#" ]]; then
        export $line
    fi
done
go run main.go
