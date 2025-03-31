from app.db.database import create_db_and_tables


def init_db():
    create_db_and_tables()


if __name__ == "__main__":
    init_db()
    print("Database tables created") 