# Bravetools Configuration Language

BCL is a simplified configuration script for Bravetools Images.

<!-- TOC -->
1. [Installation](#installation)
2. [Grammar](#grammar)
3. [Features](#features)
4. [Data Types](#datatypes)
5. [Usage](#usage)
<!-- /TOC -->

<a id="markdown-installation" name="installation"></a>
## Installation

```bash
git clone https://github.com/beringresearch/bcl
cd bcl
go get
go build
cp bcl /usr/local/bin
```
<a id="markdown-grammar" name="grammar"></a>
## Grammar

The minimal structural unit of BCL is an **Entry**. Each **Entry** is comprised of functional __Blocks__.
BCL supports five entry types:

* **base** - describes base requirements for your image, such as base image and location of the image file.
```python
base {
  image: 	"alpine/edge/amd64"
  location:     "public"
}
```

* **system** - describes system packages to be installed through a specified package manager.
Supported package managers are `atp` and `apk.
`
```python
// Install bash and python3 using apk manager
system {
    apk:  ["bash", "python3"]
}
```

* **copy** - is specialised entity designed for file and directory transfers between hosts and Brave Images.
The Entity supports multible Blocks. Each Block must be prefaced with a `key`(e.g. `Bravefile {...}`) and
contain a `source` and a `target`. Optionally, `action` specifies additional actions to perform once the file
or directory has been copied to the image. All actions are executed on an image during build. 

```python
copy {
	Bravefile {
	source:			"Bravefile.bcl"
	target: 		"/root/Bravefile.bcl"
	action: 	 	"chmod 0700 /root/Bravefile.bcl"
	}

	Bravefile {
	source:			"Bravefile"
	target: 		"/root/Bravefile"
	action: 	 	"chmod 0700 /root/Bravefile"
	}

}
```
* **run** - executes commands on the Brave image during build time. This Entity supports multiple Blocks and a diverse range of syntax.
In its simplest embodiment, `run` Entity supports command, followed by an argument string. For example,

```python
run {
  git: "clone https://github.com/beringresearch/bcl"
}
```
Complex strings are passed as Arrays.
```python
run {
  echo: ["\"clone https://github.com/beringresearch/bcl\""]
}
```

This will print an output:
```bash
$ echo "clone https://github.com/beringresearch/bcl"
$ clone https://github.com/beringresearch/bcl
```
It is also possible to use Arrays to pass multiple complex commands.

```python
run {
  bash: ["-c", "curl -sL https://bootstrap.pypa.io/get-pip.py | sudo -E python3.6"]
}
```

Multi-line strings are also supported:

```python
run {
  echo: `\"This will generate a
multiline outout
to the terminal\"` 
}
```


* **service** - controls image properties, such as name, version, and run-time configuration.

```python
service {
	name:			"alpine"
	version:		"1.0"
	ip: 			"10.0.0.1"
	ports: 			"8008:8008"

	// Commands and actions to be run after unit is deployed
	postdeploy {
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
	}

	resources {
		ram: 		"4GB"
		cpu: 		2
		gpu:		true
	}
}
```

<a id="markdown-features" name="features"></a>
## Features

### Readability
BCL parser supports arbitrary TAB and SPACE placements. This can greatly improve readability:

```python
system {
    apk:  ["bash", "python3",
           "htop", "curl"]
}
```

### Comments
Comments are designated as `//` and are natively supported by the BCL parser.

```python
// This entire entry will be ignored
//system {
//    apk: ["bash", "python3",
//          "htop", "curl"]
//}
```

<a id="markdown-usage" name="usage"></a>
## Usage 

Assuming that you have generated a BCL file `Bravefile.bcl`, to convert it to a conventional `Bravefile` run:

``` bash
$ bcl Bravefile.bcl
```

The output will generate a `Bravefile` in the working directory.
