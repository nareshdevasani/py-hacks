import requests
import subprocess

# Step 1: Query the URL and loop over the lines
url = "https://rubygems.org/versions"
response = requests.get(url)
lines = response.content.decode().splitlines()

for i in range(150, 200):
    line = lines[i]
    print("*****************************" + line)
    # Skip empty lines
    if not line:
        continue
    
    package_name = line.strip()

    # Split the line by space and get the second value
    second_value = package_name.split()[1]
    name = package_name.split()[0]

    # Split the second value by comma and get the first and last values
    versions = second_value.split(",")
    new_package_name = "gem://{}@{}".format(name, versions[len(versions) - 1])

    # Pass the new package name to the command
    cmd = ["bazel", "run", "src/golang/internal.endor.ai/x/pluginclient", "--", "--language=ruby", "--pkg={}".format(new_package_name), "--all=true"]
    subprocess.run(cmd, check=True)
