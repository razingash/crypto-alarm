import httpx

from datetime import datetime

from core.config import INTERNAL_SERVER_API
from core.logger import custom_logger


async def send_triggered_formulas(formulas: list, message: str = ""):
    """
    sends a list of formulas that users should be sent to users\n
    1 - default successful operation | 2 - deleted variable from endpoint | 3 - deleted endpoint
    """
    try:
        async with httpx.AsyncClient(timeout=10) as client:
            response = await client.post(
                url=f"{INTERNAL_SERVER_API}/notifications/push",
                json={"formulas": formulas, "message": message}
            )

        if response.status_code == 200:
            return True
        else:
            custom_logger.log_with_path(
                level=1,
                msg=f"Error during saving a list of successful triggers in crypto-gateway service at {datetime.now()}",
                filename="ExternalErrors.log"
            )

    except httpx.RequestError:
        custom_logger.log_with_path(
            level=1,
            msg=f"Error during sending a list of successful triggers on crypto-gateway service at {datetime.now()}",
            filename="ExternalErrors.log"
        )
    except Exception as e:
        print('Unpredictable error ', e)
