# Kaptaind CLI
This is a CLI written in Go for the kaptaind Broker server.
To learn more about kaptaind, visit [here].

## Getting started
Create the .kap directory and cofig file:

```
mkdir /.kap
touch config
```

Edit the config file with the following json:
```
{
 "brokerUrl": "<BROKER-URL-ADDRESS>"
}
```

Insert the correct address of the kaptaind Broker.

## Usage
### Get clusters
```
./kaptaind get clusters
```

### Get tasks
```
./kaptaind get tasks
```

### New import task
```
./kaptaind run task --sourceClusterId="mySourceCluster" --targetClusterId="myTargetCluster"
```

### delete task
```
./kaptaind delete task <task-id>
```

## Development
build using the build script.
```
 ./build.sh
```

[here]: https://github.com/kaptaind/kaptaind