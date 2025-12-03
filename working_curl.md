curl -k -u salesforce_automation:Splunk@123 -X POST "https://cisco-lcpops.splunkcloud.com:8089/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account?output_mode=json" -d "name=salesforce_UAT_test" -d "endpoint=yourinstance.my.salesforce.com" -d "sfdc_api_version=64.0" -d "auth_type=basic" -d "username=your_salesforce_username" -d "password=your_salesforce_password" -d "token=your_salesforce_security_token"


curl -k -u salesforce_automation:Splunk@123 -X POST "https://cisco-lcpops.splunkcloud.com:8089/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object?output_mode=json" -d "name=my_salesforce_input_testing" -d "account=salesforce_UAT" -d "object=Account" -d "object_fields=Id,Name,CreatedDate,LastModifiedDate" -d "order_by=LastModifiedDate" -d "start_date=2024-01-01T00:00:00.000Z" -d "interval=300" -d "delay=60" -d "index=salesforce_uat"






curl -k -u test:test@2025 -X POST "https://prd-p-l1tsd.splunkcloud.com:8089/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account?output_mode=json" -d "name=salesforce_UAT_curl" -d "endpoint=login.salesforce.com" -d "sfdc_api_version=64.0" -d "auth_type=basic" -d "username=minarva.devi777@relanto.ai" -d "password=Itsmeminu@2" -d "token=QCYk6UxyNUeYdJkRZniUJVJ3o"


curl -k -u rajshetty727@gmail.com:Password@2025 -X POST "http://127.0.0.1:8089/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account?output_mode=json" -d "name=salesforce_UAT_curl" -d "endpoint=login.salesforce.com" -d "sfdc_api_version=64.0" -d "auth_type=basic" -d "username=minarva.devi777@relanto.ai" -d "password=Itsmeminu@2" -d "token=QCYk6UxyNUeYdJkRZniUJVJ3o"


 curl -k -X POST "https://localhost:8089/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account?output_mode=json"   -H "Authorization: Splunk ^GGnvF6Vi5WAekMjSggP1a1aLTz9Vx71^_DsGmHkCIP3JCM^vP89OvvKB5NMM5aTgDYHKd5pL3c393GoBRvrVOW7dGP3o5IV5VJVi69tYN"   -d "name=salesforce_UAT_curl"   -d "endpoint=login.salesforce.com"   -d "sfdc_api_version=64.0"   -d "auth_type=basic"   -d "username=minarva.devi777@relanto.ai"   -d "password=Itsmeminu@2"   -d "token=QCYk6UxyNUeYdJkRZniUJVJ3o"


 curl -k -H "Authorization: Splunk 9oXKMsMakb7noHN4vimfyy54xIHOKQGIlgeQH8HA19meAtwG5KKu5813QWwXBeBGCa0YwZo64ECuPR^anpRG_TmiuzZ65mpeFo" -X POST "https://localhost:8089/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object?output_mode=json" -d "name=my_salesforce_input_testing_curl" -d "account=salesforce_UAT_curl" -d "object=Account" -d "object_fields=Id,Name,CreatedDate,LastModifiedDate" -d "order_by=LastModifiedDate" -d "start_date=2024-01-01T00:00:00.000Z" -d "interval=300" -d "delay=60" -d "index=default"
