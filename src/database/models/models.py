from sqlalchemy import Column, String
from src.database.connection import Base
from src.database.models.base_model import CRUDMixin


class NewsItem(Base, CRUDMixin):
    __tablename__ = "data_news_items"

    title = Column(String)
    content = Column(String)
    source = Column(String)
    # id, created_at 由 Mixin 自动提供