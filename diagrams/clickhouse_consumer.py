"""
ClickHouse Consumer service flow diagram.
Shows the micro-batch pattern: read from Kafka, buffer, flush to ClickHouse.
Run: python3 diagrams/clickhouse_consumer.py
Output: diagrams/output/clickhouse_consumer.png
"""

import os
os.chdir(os.path.join(os.path.dirname(__file__), "output"))

from diagrams import Diagram, Edge, Cluster
from diagrams.onprem.queue import Kafka
from diagrams.onprem.database import ClickHouse
from diagrams.programming.language import Go
from diagrams.onprem.inmemory import Redis

with Diagram(
    "ClickHouse Consumer — Micro-Batch Flow",
    filename="clickhouse_consumer",
    show=False,
    direction="LR",
):
    kafka = Kafka("Kafka\npageview-events\nGroupID: clickhouse-consumer")

    with Cluster("ClickHouse Consumer"):
        reader = Go("Kafka Reader\ngoroutine")
        buffer = Redis("In-Memory Buffer\n[]PageviewEvent")
        flusher = Go("Flush Logic\n5000 events OR 2s ticker")
        inserter = Go("insertBatch()\nPrepareBatch + Send")

    db = ClickHouse("ClickHouse\nanalytics.pageviews")

    kafka >> Edge(label="ReadMessage()") >> reader
    reader >> Edge(label="append to slice") >> buffer
    buffer >> Edge(label="len >= 5000\nor ticker fires") >> flusher
    flusher >> Edge(label="batch of N events") >> inserter
    inserter >> Edge(label="single INSERT\n1 part on disk") >> db
