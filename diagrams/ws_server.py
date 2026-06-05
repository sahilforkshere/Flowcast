"""
WebSocket Server service flow diagram.
Shows how stats are queried from ClickHouse and pushed to all clients.
Run: python3 diagrams/ws_server.py
Output: diagrams/output/ws_server.png
"""

import os
os.chdir(os.path.join(os.path.dirname(__file__), "output"))

from diagrams import Diagram, Edge, Cluster
from diagrams.onprem.database import ClickHouse
from diagrams.onprem.client import User
from diagrams.programming.language import Go
from diagrams.onprem.network import Haproxy

with Diagram(
    "WebSocket Server — Push Flow",
    filename="ws_server",
    show=False,
    direction="LR",
):
    browsers = User("Connected Browsers\n(N clients)")

    with Cluster("WebSocket Server :8081"):
        upgrader = Haproxy("HTTP Upgrader\nGET /ws")
        registry = Go("Client Registry\nmap[*Conn]bool\n+ sync.Mutex")
        ticker = Go("2s Ticker\ngoroutine")

        with Cluster("ClickHouse Queries"):
            q1 = Go("active users\nlast 5 min")
            q2 = Go("top 5 pages\nlast 1 hour")
            q3 = Go("views/minute\nlast 30 min")

        assembler = Go("DashboardStats{}\njson.Marshal()")

    db = ClickHouse("ClickHouse\nanalytics.pageviews")

    browsers >> Edge(label="WS handshake") >> upgrader
    upgrader >> Edge(label="register client") >> registry
    ticker >> q1
    ticker >> q2
    ticker >> q3
    q1 >> Edge(label="query") >> db
    q2 >> Edge(label="query") >> db
    q3 >> Edge(label="query") >> db
    q1 >> assembler
    q2 >> assembler
    q3 >> assembler
    assembler >> Edge(label="WriteMessage()\nJSON every 2s") >> registry
    registry >> Edge(label="push to all clients") >> browsers
