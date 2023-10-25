go build -o ./api/admin/ferrum-admin.exe ./api/admin/cli

./api/admin/ferrum-admin.exe --resource=realm --namespace=ferrum_2 --operation=get  --resource_id=IRP --value=\"{\"name\" "irp_rt\"}\"

get:
./api/admin/ferrum-admin.exe --namespace=ferrum_1 --operation=get --params=myApp --resource=client --resource_id=test-service-app-client

create:
./api/admin/ferrum-admin.exe --namespace=ferrum_1 --operation=create --resource=realm --resource_id=IRP --value='{"name":"irp_rt"}'


