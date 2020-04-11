# Bering Configuration Language

BCL is a simplified configuration script for Bravetools Images. It supports

* Comments
* Strings
* Integers
* Floats
* Boolean
* Arrays

# Example
``` bash
go get
go run main.go Bravefile.bcl
brave build Bravefile
```

# Building from source
```bash
git clone https://github.com/beringresearch/bcl
cd bcl
go get
go build
cp bcl /usr/local/bin

bcl Bravefile.bcl
```

