# grpc client server transfering images (any file possible)

# to run the project
git clone https://github.com/Chilibonitsu/imageStorage

In the imageStorage directory:

go mod tidy

make build

# in the database there are check if file was uploaded

So, after the first run, the program says "file exists" when transferring the same files, which is predictable behavior.

To transfer the same files continuously, you may comment out the section s.storage.FindFileByName(fileName) between the // //

see func (s *serverAPI) Upload(stream pb.GuploadService_UploadServer) error 

in the imageStorage\internal\grpc\serverStorage\server.go

Docker is untested.
















