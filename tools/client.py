import argparse
import requests
import json
import base64


def decode_jwt(token):
    parts = token.split('.')
    if len(parts) != 3:
        print("Invalid JWT token.")
        return

    payload = parts[1]
    padding = len(payload) % 4
    payload += "=" * (4 - padding)
    decoded = base64.urlsafe_b64decode(payload)
    return json.loads(decoded)

def get_token():
    try:
        with open(".auth.token", "r") as file:
            return file.read().strip()
    except FileNotFoundError:
        print("Token file not found. Please login first.")
        return None

def login(email, password):
    url = "http://localhost:4000/login"
    data = json.dumps({"email": email, "password": password})
    headers = {'Content-Type': 'application/json'}
    response = requests.post(url, data=data, headers=headers)

    if response.status_code == 200:
        token = response.json().get('token')
        with open(".auth.token", "w") as file:
            file.write(token)
        print("Login successful, token cached.")
        decoded_jwt_content = decode_jwt(token)
        print("JWT contents:", json.dumps(decoded_jwt_content, indent=2))
    else:
        print("Login failed.")

def signup(email, password, role):
    url = "http://localhost:4000/signup"
    data = json.dumps({"email": email, "password": password, "role": role})
    headers = {'Content-Type': 'application/json'}
    response = requests.post(url, data=data, headers=headers)

    if response.headers['Content-Type'] == 'application/json':
        print(response.json())
    else:
        print(response.text)

def change_role(email, newRole):
    token = get_token()
    if not token:
        return

    url = f"http://localhost:4000/changeRole?token={token}"
    data = json.dumps({"email": email, "newRole": newRole})
    headers = {'Content-Type': 'application/json'}
    response = requests.post(url, data=data, headers=headers)

    if response.headers['Content-Type'] == 'application/json':
        print(response.json())
    else:
        print(response.text)

def list_users():
    token = get_token()
    if not token:
        return

    url = f"http://localhost:4000/listUsers?token={token}"
    response = requests.get(url)

    if response.status_code == 200:
        print(json.dumps(response.json(), indent=2))
    else:
        print("Failed to fetch tasks. Status code:", response.status_code)

def complete_task(task_id):
    token = get_token()
    if not token:
        return

    url = f"http://localhost:4001/tasks/complete/{task_id}?token={token}"
    response = requests.post(url)

    if response.get('Content-Type') == 'application/json':
        print(response.json())
    else:
        print(response.text)

def list_tasks():
    token = get_token()
    if not token:
        return

    url = f"http://localhost:4001/tasks?token={token}"
    response = requests.get(url)

    if response.status_code == 200:
        print(json.dumps(response.json(), indent=2))
    else:
        print("Failed to fetch tasks. Status code:", response.status_code)

def shuffle_tasks():
    token = get_token()
    if not token:
        returnp

    url = f"http://localhost:4001/tasks/shuffle?token={token}"
    response = requests.post(url)

    if response.status_code == 200:
        print(json.dumps(response.json(), indent=2))
    else:
        print("Failed to shuffle tasks. Status code:", response.status_code)

def create_task(description):
    token = get_token()
    if not token:
        return

    url = f"http://localhost:4001/tasks/new?token={token}"
    data = json.dumps({"description": description})
    headers = {'Content-Type': 'application/json'}
    response = requests.put(url, data=data, headers=headers)

    if response.status_code >= 200 and response.status_code < 300:
        print("Task created successfully.")
        if response.headers.get('Content-Type') == 'application/json':
            print(response.json())
    else:
        print("Failed to create task. Status code:", response.status_code)


def main():
    parser = argparse.ArgumentParser(description="Interact with task API.")
    subparsers = parser.add_subparsers(dest="command", help="Commands")

    login_parser = subparsers.add_parser("login", help="Login to the system")
    login_parser.add_argument("email", type=str, help="Email address for login.")
    login_parser.add_argument("password", type=str, help="Password for login.")

    signup_parser = subparsers.add_parser("signup", help="Signup to the system")
    signup_parser.add_argument("email", type=str, help="Email address for signup.")
    signup_parser.add_argument("password", type=str, help="Password for signup.")
    signup_parser.add_argument("role", type=str, help="Role for signup.", default="worker")

    change_role_parser = subparsers.add_parser("change_role", help="Change user's role")
    change_role_parser.add_argument("email", type=str, help="User's email address")
    change_role_parser.add_argument("role", type=str, help="New role", default="worker")

    list_users_parser = subparsers.add_parser("list_users", help="List all users.")

    complete_task_parser = subparsers.add_parser("complete_task", help="Complete a task with given task ID.")
    complete_task_parser.add_argument("task_id", type=str, help="The ID of the task to complete.")

    list_tasks_parser = subparsers.add_parser("list_tasks", help="List all tasks.")

    shuffle_tasks_parser = subparsers.add_parser("shuffle_tasks", help="Shuffle all tasks.")

    create_task_parser = subparsers.add_parser("create_task", help="Create a new task with a description.")
    create_task_parser.add_argument("description", type=str, help="Description of the new task.")

    args = parser.parse_args()

    if args.command == "login":
        login(args.email, args.password)
    elif args.command == "signup":
        signup(args.email, args.password, args.role)
    elif args.command == "list_users":
        list_users()
    elif args.command == "change_role":
        change_role(args.email, args.role)
    elif args.command == "complete_task":
        complete_task(args.task_id)
    elif args.command == "list_tasks":
        list_tasks()
    elif args.command == "create_task":
        create_task(args.description)
    elif args.command == "shuffle_tasks":
        shuffle_tasks()

if __name__ == "__main__":
    main()
