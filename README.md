<h1 align="center">Hi there, I'm <a href="https://github.com/ogamor69" target="_blank">Roman</a> 

<h3 align="center">Affise Task Fetcher</h3>



Affise Task Fetcher is a small Go application that fetches data from a list of given URLs concurrently and returns the fetched data in JSON format. The application includes a limited listener, which allows a limited number of simultaneous connections, and a handler that processes the list of URLs and performs HTTP requests to each of them.

#Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Testing](#testing)


##Installation

1.Make sure you have Go installed on your machine. You can check the installation by running go version. If it's not installed, follow the instructions on Go's official website.
2.Clone the repository:
```bash 
git clone https://github.com/ogamor69/affise.git
```
3.Change into the project directory:
```bash 
cd affise
```
4.Build the application:
```bash 
go build
```


##Usage

1.Run the application:
```bash 
./affise
```
2.The server will start listening on :8080.
3.Send a POST request with a JSON payload containing a list of URLs to fetch:
```bash 
curl -X POST -H "Content-Type: application/json" -d '{"urls": ["https://www.google.com", "https://www.yahoo.com"]}' http://localhost:8080
```

##Testing

1.Run the tests:
```bash 
go test
```

