import os
from sqlmodel import SQLModel, create_engine, Session

DATABASE_URL = os.getenv("DATABASE_URL", "sqlite:///./orders.db")
engine = create_engine(
    "sqlite:///./orders.db", 
    connect_args={"check_same_thread": False} # Avoid SQLite thread safety model
)


def create_db_and_tables():
    SQLModel.metadata.create_all(engine)


def get_db_session():
    with Session(engine) as session:
        yield session 