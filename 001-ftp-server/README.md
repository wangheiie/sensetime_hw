# Only Myftp Server
> This project is a simple for practis training, written in Golang.**Welcome criticism and corrction. I feel honored to learn from your help** .

## Platform
There are three files `myftp.go`,`Dockerfile`and`nyftp-k8s.yml`. Build docker images, and details are in `Dockerfile`. Maybe you can build it with `docker build -t doublered/myftp:versager`. To deploy it on K8s, the template is `myftp-k8s.yml`. Maybe you can deploy with `kubectl create -f myftp-k8s.yml`.  

## Usage
* Really advise do not use this project in any case, because it has not been tested so far. It is only for training.
* You can use `--help` to see the usage info.
* `myftp.go` has three parameters, host, port and dir, which assigns where the server listening and where the files be.

## Support Commands
Now supports the following commands temporarily.
* USER/PASS: Username and password check
* PWD: get current path
* CWD: change current path
* LIST: list the cotent of path
* PASV: Enter passive node to transfer files
* RETR: return files form server
* STOR: store files to server
* PORT: maybe you have seen this command but that is your heteroptics. this is not support. 

## Commit-logs
* 2018-03-26: Add project.

# License
Maybe there will be under some license. But it is a temporary vacancy.
