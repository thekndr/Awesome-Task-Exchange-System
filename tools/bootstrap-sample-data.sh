#!/usr/bin/env sh

set -ex

python3 client.py signup worker-one@uberpopug.com password-one worker
python3 client.py signup worker-two@uberpopug.com password-two worker
python3 client.py signup another-worker@uberpopug.com password-three worker

python3 client.py login worker-one@uberpopug.com password-one

python3 client.py change_role another-worker@uberpopug.com manager

python3 client.py login another-worker@uberpopug.com password-three

python3 client.py create_task "task 0"
python3 client.py create_task "task 1"
python3 client.py create_task "task 2"
python3 client.py create_task "task 3"

task_0_id="$(python3 client.py list_tasks | jq -r '.[] | select(.description == "task 0") | .id')"
python3 client.py complete_task "${task_0_id}"

python3 client.py shuffle_tasks
