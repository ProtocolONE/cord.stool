# Push command
   Upload update app bundle to one of servers

## Usage
```sh
   cord.stool.exe push [command options] [arguments...]
```

## Options
```sh
   --source value, -s value  Path to game
   --output value, -o value  Path to upload
   --ftp value               Full ftp url path. Example ftp://user:password@host:port/upload/directory
   --sftp value              Full sftp url path. Example sftp://user:password@host:port/upload/directory
   --aws-region value        AWS region name
   --aws-credentials value   Path to AWS credentials file
   --aws-profile value       AWS profile name
   --aws-id value            AWS access key id
   --aws-key value           AWS secret access key
   --aws-token value         AWS session token
   --s3-bucket value         Amazon S3 bucket name
   --akm-hostname value      Akamai hostname
   --akm-keyname value       Akamai keyname
   --akm-key value           Akamai key
   --akm-code value          Akamai code
   --cord-url value          Cord server url
   --cord-login value        Cord user login
   --cord-password value     Cord user password
   --cord-patch              Upload the difference between files using xdelta algorithm
   --cord-hash               Upload changed files only
   --cord-wsync              Upload changed files only using Wharf protocol that enables incremental uploads   
```

## Description
   Command **push** uploads files to one of specified network storage.</br>
   Use option **--source** to specify path to files to be uploaded and option **--output** to specify remote relative path to store the files to be uploaded  (optional).</br>
### There are options specific for the each network storage
   Use option **--ftp** to specify a full ftp url path, example ftp://user:password@host:port/upload/directory.</br>
   Use option **--sftp** to specify a full sftp url path, example sftp://user:password@host:port/upload/directory.</br>
#### Amazon Simple Storage Service (Amazon S3)
   Use option **--aws-region** to specify a AWS region name.</br>
   Use option **--aws-credentials** to specify a path to AWS credentials file.</br>
   Use option **--aws-profile** to specify a AWS profile name (optional).</br>
   Use option **--aws-id** to specify a AWS access key id (applying if aws-credentials is not specified).</br>
   Use option **--aws-key** to specify a AWS secret access key (applying if aws-credentials is not specified).</br>
   Use option **--aws-token** to specify a AWS session token (applying if aws-credentials is not specified, optional).</br>
   Use option **--s3-bucket** to specify a Amazon S3 bucket name.</br>
#### Akamai CDN
   Use option **--akm-hostname** to specify Akamai hostname.</br>
   Use option **--akm-keyname** to specify Akamai keyname.</br>
   Use option **--akm-key** to specify Akamai key.</br>
   Use option **--akm-code** to specify Akamai code.</br>
#### Cord Server
   Use option **--cord-** to specify Cord server hostname.</br>
   Use option **--cord-login** to specify Cord server login.</br>
   Use option **--cord-password** to specify Cord server password.</br>
   Use option **--cord-patch** to specify to upload the difference between files only. It uses xdelta algorithm.</br>
   Use option **--cord-hash** to specify to upload changed files only. It uses files hash to find out changed files.</br>
   Use option **--cord-wsync** to specify to upload changed files only. It uses Wharf protocol that enables incremental uploads.</br>
