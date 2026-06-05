"""
Event Producer service flow diagram.
Shows how a single POST /track request is handled.
Run: python3 diagrams/event_producer.py
Output: diagrams/output/event_producer.png
"""

import os
os.chdir(os.path.join(os.path.dirname(__file__), "output"))

from diagrams import Diagram, Edge, Cluster
from diagrams.onprem.queue import Kafka
from diagrams.onprem.client import User
from diagrams.programming.language import Go
from diagrams.onprem.network import Haproxy

with Diagram(
    "Event Producer — Request Flow",
    filename="event_producer",
    show=False,
    direction="LR",
):
    client = User("Client\n(browser / k6)")

    with Cluster("Event Producer :8080"):
        router = Haproxy("HTTP Router")
        validator = Go("JSON Validator\ncheck session_id\ncheck page_url")
        writer = Go("Kafka Writer\nBatchSize: 100\nBatchTimeout: 10ms\nKey: session_id")

    kafka = Kafka("Kafka\npageview-events")

    client >> Edge(label="POST /track\n{session_id, page_url, ...}") >> router
    router >> Edge(label="valid request") >> validator
    validator >> Edge(label="202 Accepted\n(immediate)") >> client
    validator >> Edge(label="serialize to JSON") >> writer
    writer >> Edge(label="write message\nkey=session_id") >> kafka
