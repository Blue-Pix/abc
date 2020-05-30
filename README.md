# Overview
`abc` is a helper library including several sub commands.  
It wraps, pipes and extends AWS API.  

## Table of Contents

- [Install](#install)
- [Credentials and Permissions](#credentials-and-permissions)
- [Usage](#usage)
  - [abc ami](#abc-ami)
  - [abc cfn unused-exports](#abc-cfn-unused-exports)
- [License](#license)
- [Contributing](#contributing)

# Install
You can install binaries from [releases](https://github.com/Blue-Pix/abc/releases),   
or pull [repository](https://github.com/Blue-Pix/abc) and build the binaries yourself.

If you build yourself:  
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

# Credentials and Permissions

You have to set aws credentials in advance.  
Which policy to use depends on sub command type.  
  
For example, `abc ami` command querying latest Amazon Linux AMI, requires following policy.

- ssm:GetParametersByPath

Without option, we use your default aws credentials.  
You can also pass `--region` and `--profile` option like aws cli.

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

Originally, it returns 10~15 AMIs, as parameter path is `/aws/service/ami-amazon-linux-latest` and search sub directory recursively.  
If you wanna spare time to find the path of the AMI, use this helper and query it!

## `abc cfn unused-exports`

List Cloudformation's exports, which not used in any stack.  
It prints `name` and `exporting_stack` as csv with header.

**Example:**

There is a stack named `abc-sample-stack` with following template.

```yaml
AWSTemplateFormatVersion: "2010-09-09"
Parameters:
  PJ:
    Description: project identifier
    Type: String
    Default: abc
Resources:
  Queue1:
    Type: AWS::SQS::Queue
    Properties: 
      QueueName: !Sub ${PJ}-queue1
  Queue2:
    Type: AWS::SQS::Queue
    Properties: 
      QueueName: !Sub ${PJ}-queue2
Outputs:
  Queue1:
    Value: !GetAtt Queue1.Arn
    Export:
      Name: !Sub ${PJ}-queue1-arn
  Queue2:
    Value: !GetAtt Queue2.Arn
    Export:
      Name: !Sub ${PJ}-queue2-arn
```

```sh
$ abc cfn unused-exports
[{"name":"abc-queue1-arn","exporting_stack":"abc-sample-stack"},{"name":"abc-queue2-arn","exporting_stack":"abc-sample-stack"}]
```

# License
This code is made available under the Apache License 2.0.

# Contributing
Feel free to open issues 🎉
