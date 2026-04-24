import os
import re

for filename in os.listdir("internal/provider/"):
    if filename.endswith("_test.go"):
        with open("internal/provider/" + filename, "r") as f:
            content = f.read()
        
        # Replace 'c := hyperping.NewClient("test_api_key")'
        # with 'clients := &hyperpingClients{REST: hyperping.NewClient("test_api_key")}'
        content = re.sub(r'c := hyperping.NewClient\("test_api_key"\)', 'clients := &hyperpingClients{REST: hyperping.NewClient("test_api_key")}', content)
        
        # Replace 'ProviderData: c,' or 'ProviderData: c'
        content = re.sub(r'ProviderData:\s+c([,}])', r'ProviderData: clients\1', content)
        
        with open("internal/provider/" + filename, "w") as f:
            f.write(content)
