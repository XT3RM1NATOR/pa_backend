package storage


minioClient, err := minio.New(endpoint, &minio.Options{
Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
Secure: useSSL,
})
if err != nil {
log.Fatalln("Failed to create MinIO client:", err)
}

// Test connection
ctx := context.Background()
found, err := minioClient.BucketExists(ctx, "test-bucket")
if err != nil {
log.Fatalln("Error checking bucket existence:", err)
}
if found {
fmt.Println("Bucket 'test-bucket' exists!")
} else {
fmt.Println("Bucket 'test-bucket' does not exist.")
}
