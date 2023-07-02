from flask import Flask, jsonify, make_response, request
import urllib3

import uuid
import datetime
import json
import hashlib
import base64
import borg

import subprocess
app = Flask(__name__)
db = borg.DB()
http = urllib3.PoolManager()
@app.route("/")
def hello_from_root():
    return jsonify(message='Hello from root!')


@app.route("/hello")
def hello():
    return jsonify(message='Hello from path! 2')

@app.route('/api/total', methods=['POST'])
def endpoint():
    request_data = request.get_json()

    # Extracting values from the request
    request_id = request_data['requestId']
    request_time = request_data['requestTime']
    value1 = request_data['data']['value1']
    value2 = request_data['data']['value2']

    # Perform operations or calculations with the values
    result = value1 + value2

    # Prepare the response
    response = {
        "requestId": uuid.uuid4(),
        "requestTime":  datetime.datetime.utcnow(),
        "input": json.dumps(request_data),
        "result": result,
    }

    return jsonify(response)

@app.route('/api/signature', methods=['POST'])
def signature():
    request_data = request.get_json()

    # Generate unique requestId using UUID
    request_id = str(uuid.uuid4())

    # Get the current time as requestTime
    request_time = datetime.datetime.now().isoformat()
    print(request_data)
    # Extracting values from the request
    p = request_data['data']['plaintText']
    secret_key = request_data['data']['secretKey']

    # Generate the signature using plaintext and secret_key
    s = hashlib.sha256((p + secret_key).encode()).hexdigest()

    # Prepare the response
    response = {
        "requestId": request_id,
        "requestTime": request_time,
        "signature": s
    }

    return jsonify(response)

@app.route('/api/base64', methods=['POST'])
def base64_func():
    request_data = request.get_json()

    # Generate unique requestId using UUID
    request_id = str(uuid.uuid4())

    # Get the current time as requestTime
    request_time = datetime.datetime.now().isoformat()

    # Extracting values from the request
    need_encode = request_data['data']['needEncode']
    need_decode = request_data['data']['needDecode']

    # Base64 encode the need_encode string
    encoded_string = base64.b64encode(need_encode.encode()).decode()

    # Base64 decode the need_decode string
    decoded_string = base64.b64decode(need_decode).decode()

    # Prepare the response
    response = {
        "requestId": request_id,
        "requestTime": request_time,
        "encodedString": encoded_string,
        "decodedString": decoded_string
    }

    return jsonify(response)


@app.route('/api/user', methods=['POST'])
def insert_user():
    request_data = request.get_json()

    # Extract data from the request
    username = request_data['data']['username']
    name = request_data['data']['name']
    phone = request_data['data']['phone']

    # Check the uniqueness of the username in the database
    if is_username_unique(username):
        # Create a new user in the database
        cursor = db.conn.cursor()
        insert_query = '''
            INSERT INTO users_phuc (username, name, phone)
            VALUES (%s, %s, %s)
        '''
        cursor.execute(insert_query, (username, name, phone))
        db.conn.commit()
        cursor.close()

        # Create response
        response = {
            'responseId': request_data['requestId'],
            'responseTime': datetime.datetime.now().isoformat(),
            'responseCode': '200',
            'responseMessage': 'User created successfully'
        }
    else:
        response = {
            'responseId': request_data['requestId'],
            'responseTime': datetime.datetime.now().isoformat(),
            'responseCode': '400',
            'responseMessage': 'Username already exists'
        }

    return jsonify(response)


def is_username_unique(username):
    # Check the uniqueness of the username in the database
    cursor = db.conn.cursor()
    select_query = "SELECT * FROM users_phuc WHERE username = %s"
    cursor.execute(select_query, (username,))
    user = cursor.fetchone()
    cursor.close()
    return user is None


@app.route('/update_user', methods=['POST'])
def update_user():
    request_data = request.get_json()

    # Extract data from the request
    username = request_data['data']['username']
    name = request_data['data']['name']
    phone = request_data['data']['phone']

    # Check and update the user in the database
    if update_user_info(username, name, phone):
        # Update successful
        response = {
            'responseId': request_data['requestId'],
            'responseTime': datetime.datetime.now().isoformat(),
            'responseCode': '200',
            'responseMessage': 'User updated successfully'
        }
    else:
        # Update failed (user not found)
        response = {
            'responseId': request_data['requestId'],
            'responseTime': datetime.datetime.now().isoformat(),
            'responseCode': '400',
            'responseMessage': 'User not found'
        }

    return jsonify(response)


def update_user_info(username, name, phone):
    # Check and update user information in the database
    cursor = db.conn.cursor()
    update_query = '''
        UPDATE users_phuc
        SET name = %s, phone = %s
        WHERE username = %s
    '''
    cursor.execute(update_query, (name, phone, username))
    db.conn.commit()
    updated_rows = cursor.rowcount
    cursor.close()
    return updated_rows > 0


@app.route('/delete_user', methods=['POST'])
def delete_user():
    request_data = request.get_json()

    username = request_data['data']['username']

    if delete_user_by_username(username):
        response = {
            'responseId': request_data['requestId'],
            'responseTime': datetime.datetime.now().isoformat(),
            'responseCode': '200',
            'responseMessage': 'User deleted successfully'
        }
    else:
        response = {
            'responseId': request_data['requestId'],
            'responseTime': datetime.datetime.now().isoformat(),
            'responseCode': '400',
            'responseMessage': 'User not found'
        }

    return jsonify(response)


def delete_user_by_username(username):
    cursor = db.conn.cursor()
    delete_query = '''
        DELETE FROM users_phuc
        WHERE username = %s
    '''
    cursor.execute(delete_query, (username,))
    db.conn.commit()
    deleted_rows = cursor.rowcount
    cursor.close()
    return deleted_rows > 0


@app.route('/user/<username>', methods=['GET'])
def get_user(username):
    user = get_user_by_username(username)

    if user is not None:
        response = {
            'responseId': str(uuid.uuid4()),
            'responseTime': datetime.datetime.now().isoformat(),
            'responseCode': '200',
            'responseMessage': 'User found',
            'data': {
                'username': {
                    'ID': user[0],
                    'username': user[1],
                    'name': user[2],
                    'phone': user[3],

                }
            }
        }
    else:
        response = {
            'responseId': str(uuid.uuid4()),
            'responseTime': datetime.datetime.now().isoformat(),
            'responseCode': '400',
            'responseMessage': 'User not found'
        }

    return jsonify(response)


@app.route('/mobile', methods=['POST'])
def check_mobile():
    # Extract data from the request
    request_id = request.json.get('requestId')
    request_time = request.json.get('requestTime')
    signature = request.json.get('signature')
    data = request.json.get('data')

    # Verify the signature
    phone = data.get('phone')
    username = data.get('username')
    secret_key = 'golang'  # Replace with your secret key
    expected_signature = hashlib.sha256((request_id + phone + username + secret_key).encode()).hexdigest()
    print(expected_signature)
    if signature != expected_signature:
        return {
            "responseId": request_id,
            "responseTime": "current_time",
            "responseCode": "INVALID_SIGNATURE",
            "responseMessage": "Invalid signature"
        }
    phone_number = data.get('phone')
    last_number = int(phone_number[-1])
    # Perform any desired operations with the data
    # ...

    # Send the request using curl
    payload = {
        'requestId': request_id,
        'data': {
            'value': last_number
        }
    }
        
    # Set the request headers
    headers = {
        'x-api-key': 'B5d4JtTU8u1ggV8gp7OF88gcCGxZls6T3f5PYZSa',
        'Content-Type': 'text/plain'
    }

    json_responseBody = json.dumps(payload)
# Make the API request
    response = http.request('POST', 'https://1g1zcrwqhj.execute-api.ap-southeast-1.amazonaws.com/dev/testapi', headers=headers, body=json_responseBody)
    print("Status code:", response.status)    
    data = json.loads(response.data)
    if data['responseCode'] != "00":
       return jsonify({
        "responseId": request_id,
        "responseTime": "current_time",
        "responseCode": "SUCCESS",
        "responseMessage": "KH HOP LE"
        })
    # Process the response
    # if response.status_code == 400:
    #     response_data = response.json()
    #     # Process the response data
    #     print(response_data)
    # else:
    #     # Handle the request error
    #     print('Request failed with status code:', response)

    # Process the response from the external API
    # ...
    # Return the response
    
    return jsonify({
        "responseId": request_id,
        "responseTime": "current_time",
        "responseCode": "SUCCESS",
        "responseMessage": "HOP LE"
    })
    
   




def get_user_by_username(username):
    cursor = db.conn.cursor()
    select_query = '''
        SELECT *
        FROM users_phuc
        WHERE username = %s
    '''
    cursor.execute(select_query, (username,))
    user = cursor.fetchone()
    cursor.close()
    return user


@app.errorhandler(404)
def resource_not_found(e):
    return make_response(jsonify(error='Not found!'), 404)
