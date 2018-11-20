module uplgen

require (
	github.com/aws/aws-sdk-go v1.15.78 // indirect
	github.com/jlaffaye/ftp v0.0.0-20181101112434-47f21d10f0ee // indirect
	github.com/stretchr/testify v1.2.2
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
	uplgen/appargs v1.0.0
	uplgen/updateapp v1.0.0
	uplgen/updater v1.0.0
	uplgen/utils v1.0.0
)

replace uplgen/appargs v1.0.0 => ./appargs

replace uplgen/updateapp v1.0.0 => ./updateapp

replace uplgen/utils v1.0.0 => ./utils

replace uplgen/updater v1.0.0 => ./updater
