# file-collector

## Overview

file-collector collects files , copies them and executes command before/after copying.

## Usage 
```
file-collector -c config.json
```

## Configuration File

Configuration File is in JSON format.

|Property|Type|Description|Required|
|--------|----|-----------|--------|
|srcs|Array of `src`|Details are later.|Yes|
|dst|string|The root directory path to copy file.|Yes|
|after_cmd|string|The command which is executed after copying. If exit code is not 0, cancel copying.|No|

### src property

|Property|Type|Description|Required|
|--------|----|-----------|--------|
|path|string|File path to copy.|Yes|
|dst_path|string|Destination path. It should be relative path. The file will be copied under `dst`.|Yes|
|checksum|string|Generate checksum file. The file name will be `path` + `.` + `checksum`. (e.g. sample.txt.md5) `md5`, `sha1` and `sha256` are supported.|No|
|before_cmd|string|The command which is executed before copying. If exit code is not 0, cancel copying. `${target}` will be replaced by `path`. |No|
|after_cmd|string|The command which is executed after copying. If exit code is not 0, cancel copying. `${target}` will be replaced by `dst_path`.|No|

## License

[Apache License v2.0](https://www.apache.org/licenses/LICENSE-2.0)
