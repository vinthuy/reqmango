import requests
import sys

BASE_URL = "http://localhost:8000/api/v1"

def test_login():
    print("Testing login...")
    
    login_data = {"email": "pgtest@example.com", "password": "123"}
    response = requests.post(f"{BASE_URL}/auth/login", json=login_data)
    
    print(f"Status code: {response.status_code}")
    print(f"Response: {response.text}")
    
    if response.status_code == 200:
        print("Login successful")
        return response.json().get('access_token')
    else:
        print("Login failed")
        return None

def test_register():
    print("Testing register...")
    
    register_data = {
        "email": "test_new@example.com",
        "username": "testuser",
        "display_name": "Test User",
        "password": "password123"
    }
    response = requests.post(f"{BASE_URL}/auth/register", json=register_data)
    
    print(f"Status code: {response.status_code}")
    print(f"Response: {response.text}")
    
    if response.status_code == 201:
        print("Register successful")
        return True
    else:
        print("Register failed")
        return False

if __name__ == "__main__":
    print("Starting login test...")
    
    # First try to login with existing user
    token = test_login()
    
    if not token:
        print("Trying to register new user...")
        if test_register():
            print("Register succeeded, now trying login again...")
            token = test_login()
    
    if token:
        print(f"Got token: {token[:20]}...")
        print("Test passed!")
    else:
        print("Test failed!")
        sys.exit(1)
