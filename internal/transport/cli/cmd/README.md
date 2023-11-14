go build -o ./api/admin/ferrum-admin.exe ./api/admin/cli

client
./api/admin/ferrum-admin.exe --params=myApp    --resource=client --operation=get       --resource_id=test-service-app-client
./api/admin/ferrum-admin.exe --params=testRalm --resource=client --operation=delete    --resource_id=test-service-app-client
./api/admin/ferrum-admin.exe --params=myApp --resource=client --operation=create --value='{"id": "d4dc483d-7d0d-4d2e-a0a0-2d34b55e6666", "name": "training", "type": "confidential", "auth": {"type": 1, "value": "fb6Z4RsOadVycQoeQiN57xpu8w8wTEST"}}'

user:
./api/admin/ferrum-admin.exe --params=myApp --resource=user --operation=get    --resource_id=admin
./api/admin/ferrum-admin.exe --params=myApp --resource=user --operation=create --value='{"info": {"sub": "667ff6a7-3f6b-449b-a217-6fc5d9actest", "email_verified": false, "roles": ["admin"], "name": "firstTestName lastTestName", "preferred_username": "testuser", "given_name": "firstTestName", "family_name": "lastTestName"}, "credentials": {"password": "1s2d3f4g90xs"}}'

realm:
./api/admin/ferrum-admin.exe --resource=realm --operation=get    --resource_id=myApp
./api/admin/ferrum-admin.exe --resource=realm --operation=create --value='{"name": "testRalm", "token_expiration": 600, "refresh_expiration": 300, "clients": [{"id": "d4dc483d-7d0d-4d2e-a0a0-2d34b55e1111", "name": "trainingFirst", "type": "confidential", "auth": {"type": 1, "value": "fb6Z4RsOadVycQoeQiN57xpu8w8wTEST"}}, {"id": "d4dc483d-7d0d-4d2e-a0a0-2d34b55e2222", "name": "trainingSecond", "type": "confidential", "auth": {"type": 1, "value": "fb6Z4RsOadVycQoeQiN57xpu8w8wTEST"}}], "users": [{"info": {"sub": "667ff6a7-3f6b-449b-a217-111111actest", "email_verified": false, "roles": ["admin"], "name": "firstTestName lastTestName", "preferred_username": "testFirst", "given_name": "firstTestName", "family_name": "lastTestName"}, "credentials": {"password": "1s2d3f4g90xs"}}, {"info": {"sub": "667ff6a7-3f6b-449b-a217-222222actest", "email_verified": false, "roles": ["admin"], "name": "firstTestName lastTestName", "preferred_username": "testSecond", "given_name": "firstTestName", "family_name": "lastTestName"}, "credentials": {"password": "1s2d3f4g90xs"}}]}'

./api/admin/ferrum-admin.exe --params=testRalm --resource=user --operation=get --resource_id=testFirst
./api/admin/ferrum-admin.exe --params=testRalm --resource=user --operation=get --resource_id=testSecond




create:
./api/admin/ferrum-admin.exe --operation=create --resource=realm --resource_id=IRP --value='{"name":"irp_rt"}'

./api/admin/ferrum-admin.exe --operation=create --resource=realm --resource_id=IRP --value='{"name":"irp_rt"}'


