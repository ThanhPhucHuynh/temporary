from flask import Flask, jsonify, make_response, request
import uuid
import datetime
import json
import hashlib
import base64
app = Flask(__name__)


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

@app.errorhandler(404)
def resource_not_found(e):
    return make_response(jsonify(error='Not found!'), 404)




# export AWS_ACCESS_KEY_ID=AKIA5QFWU3GWCD7CQWCS
# export AWS_SECRET_ACCESS_KEY=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
# export AWS_DEFAULT_REGION==us-east-1