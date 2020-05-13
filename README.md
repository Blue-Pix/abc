# Overview
`abc` is a helper library including several sub commands.  
It wraps, pipes and extends AWS API.  

# Install
Install source code with go, build the binaries yourself, and run as `abc` command.

**Prerequisition**
- Git
- Go

```sh
mkdir $HOME/src
cd $HOME/src
git clone https://github.com/Blue-Pix/abc.git
cd abc
go install
```

# Permissions

You have to set aws credentials in advance.  
Which policy to use depends on sub command type.  
  
For example, `abc ami` command querying latest Amazon Linux AMI, requires following policy.

- ssm:GetParametersByPath

# Usage
To list all sub commands, type `abc help`.

## `abc ami`

List latest Amazon Linux AMI.  
You can also query it by version, virtualization type, cpu architecture, storage type and if minimal or not.

If you looking for the AMI composed of Amazon Linux 2, hvm, x86_64, gp2:

```sh
$ abc ami -v 2 -V hvm -a x86_64 -s gp2 | jq '.'
[
  {
    "os": "amzn",
    "version": "2",
    "virtualization_type": "hvm",
    "arch": "x86_64",
    "storage": "gp2",
    "minimal": false,
    "id": "ami-0f310fced6141e627",
    "arn": "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"
  }
]
```

**Please check configured region**  

This result is in ap-northeast-1 (Asia/Tokyo).

Originally, it returns 10~15 AMIs, as parameter path is `/aws/service/ami-amazon-linux-latest` and search sub directory recursively.  
If you wanna spare time to find the path of the AMI, use this helper and query it!

# License
This code is made available under the Apache License 2.0.

# Contributing
Feel free to open issues 🎉
