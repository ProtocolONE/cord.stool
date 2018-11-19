module uplgen

require (
	github.com/aws/aws-sdk-go v1.15.78 // indirect
	github.com/jlaffaye/ftp v0.0.0-20181101112434-47f21d10f0ee // indirect
	uplgen/appargs v1.0.0
	uplgen/updateapp v1.0.0
	uplgen/utils v1.0.0
)

replace uplgen/appargs v1.0.0 => ./appargs

replace uplgen/updateapp v1.0.0 => ./updateapp

replace uplgen/utils v1.0.0 => ./utils
