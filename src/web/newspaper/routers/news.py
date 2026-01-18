from fastapi import APIRouter

router = APIRouter(prefix="/api/newspaper", tags=["Newspaper"])

@router.get("/latest")
def get_latest_news():
    return {"news": []}
