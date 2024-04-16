import mii
import time
import argparse

parser = argparse.ArgumentParser()
parser.add_argument('--uri', required=True, help='Model URI e.g. microsoft/phi-1_5')

args = parser.parse_args()

client = mii.serve(args.uri,
                   deployment_name="default",
                   enable_restful_api=True,
                   restful_api_port=8080,
                   restful_api_host="0.0.0.0")

while True:
    time.sleep(1000)
