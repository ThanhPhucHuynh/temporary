Yes, it is definitely possible to scale up storage to 100TB or more. There are several options for doing so, depending on your specific requirements and constraints. Here are a few possibilities:

docker run \
   -p 9000:9000 \
   -p 9090:9090 \
   --name minio \
   -v ~/working/minio/dat/:/data \
   -e "MINIO_ROOT_USER=TPHUC" \
   -e "MINIO_ROOT_PASSWORD=TPHUC" \
   quay.io/minio/minio server /data --console-address ":9090"