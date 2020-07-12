import requests
import json
HEADERS = {"Content-Type":"application/json"}
URL = 'http://localhost:8080/users/'
def create_user():
    payload = {'id':'1', 'name':'chenglch', 'age':16}
    data = json.dumps(payload)
    r = requests.post(URL, headers=HEADERS, data=data)
    if r.status_code >=300:
        raise
    print(r.text)

def get_user():
    r = requests.get(URL)
    if r.status_code >=300:
        raise
    print(r.text)

def remove_user():
    params={'user-id': 1}
    # below is error as url is http://localhost:8080/users/?user-id=1
    #r = requests.delete(URL, headers=HEADERS, params=params)
    r = requests.delete('%s/%s'%(URL, '1'), headers=HEADERS)
    import pdb
    pdb.set_trace()
    if r.status_code >=300:
        raise
    print(r.text)

if __name__ == '__main__':
    create_user()
    get_user()
    remove_user()
    print("After remove user\n")
    get_user()