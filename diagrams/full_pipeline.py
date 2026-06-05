"""
Full pipeline architecture diagram.
Shows the complete data flow from browser to dashboard.
Run: python3 diagrams/full_pipeline.py
Output: diagrams/output/full_pipeline.png
"""

import os
os.chdir(os.path.join(os.path.dirname(__file__), "output"))

from diagrams import Diagram, Edge, Cluster
from diagrams.onprem.queue import Kafka
from diagrams.onprem.database import ClickHouse
from diagrams.onprem.client import User
from diagrams.programming.language import Go
from diagrams.onprem.network import Nginx

with Diagram(
    "Flowcast — Full Pipeline",
    filename="full_pipeline",
    show=False,
    direction="LR",
):
    browser = User("Browser / k6")

    with Cluster("Event Ingestion"):
        producer = Go("Event Producer\nPOST /track :8080")

    with Cluster("Message Queue"):
        kafka = Kafka("Kafka\npageview-events\n3 partitions")

    with Cluster("Storage Layer"):
        consumer = Go("ClickHouse Consumer\nmicro-batch 5k rows / 2s")
        db = ClickHouse("ClickHouse\nanalytics.pageviews\nMergeTree")

    with Cluster("Live Delivery"):
        ws = Go("WebSocket Server\n/ws :8081\nquery every 2s")
        dashboard = Nginx("Next.js Dashboard\nRecharts :3000")

    browser >> Edge(label="POST /track") >> producer
    producer >> Edge(label="kafka-go\nBatchSize 100") >> kafka
    kafka >> Edge(label="consumer group\nclickhouse-consumer") >> consumer
    consumer >> Edge(label="INSERT batch") >> db
    db >> Edge(label="3 queries / tick") >> ws
    ws >> Edge(label="JSON push\nevery 2s") >> dashboard
    dashboard >> Edge(label="renders live") >> browser
