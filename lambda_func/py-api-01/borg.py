import psycopg2
import os
# from psycopg2.extras import DictCursor
from psycopg2.extras import DictCursor

psycopg2.extensions.register_type(psycopg2.extensions.UNICODE)
psycopg2.extensions.register_type(psycopg2.extensions.UNICODEARRAY)

class DB(object):
    """Borg pattern singleton"""
    __state = {}
    def __init__(self):
        self.__dict__ = self.__state
        if not hasattr(self, 'conn'):
            self.conn = psycopg2.connect(database=os.getenv("PG_DATABASE"),
                                         user=os.getenv("PG_USER"),
                                         host=os.getenv("PG_HOST"),
                                         port="5432",
                                         password= os.getenv("PG_PASS")
                                        )
            self.cur = self.conn.cursor(cursor_factory=DictCursor)