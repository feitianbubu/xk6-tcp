publish:
	GOOS=linux xk6 build v0.39.0 --with github.com/feitianbubu/xk6-tcp="C:\git\go\xk6-tcp" --output ./k6
	rsync -av ./* blacknull@172.24.140.134:/home/blacknull/xk6


proto:
 protoc --include_imports --descriptor_set_out=tmp.pb -I proto --go_out=proto --go_opt=paths=source_relative --go-grpc_out=proto --go-grpc_opt=paths=source_relative ./proto/auth/auth.proto

protoc --include_imports --descriptor_set_out=auth.proto-tmp.pb -I proto proto/auth/auth.proto