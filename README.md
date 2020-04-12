# Bering Configuration Language

BCL is a simplified configuration script for Bravetools Images. It supports

* Comments
* Strings
* Integers
* Floats
* Boolean
* Arrays

# Building from source

```bash
git clone https://github.com/beringresearch/bcl
cd bcl
go get
go build
cp bcl /usr/local/bin
```

# Examples

## Generate BCL template file

```bash
bcl

//BCL File example
base {
	image: 		""
	location: 	""
}

system {
    apt: 		["bash", "python3"]
}

copy {
	FileName {
	source:		"/file/or/directory"
	target: 	"/file/or/directory"
	action: 	 "chmod 0700 /file/or/directory"
	}
}

run {
	echo: 		"Hello World"
}

service {
	name:		""
	version:	"1.0"
	ip: 		""
	resources {
		ram: 	"4GB"
		cpu: 	2
		gpu:	false
	}
}
```

To save output into a file, simply pipe it.

```bash
bcl > Bravefile.bcl
```

## Convert BCL file to Bravefile

```
bcl Bravefile.bcl
```

The output will generate a `Bravefile` in the working directory.