import sys

import requests

response = requests.get("http://localhost:80")
if(response.status_code == 404):
    print ("Success")
    sys.exit(0)
else:
    print ("Fail")
    sys.exit(1)
