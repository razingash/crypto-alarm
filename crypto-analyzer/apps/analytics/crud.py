from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from core.models import TriggerFormula
from db.postgre import postgres_db


async def get_formula_by_id(session: AsyncSession, pk: int) -> str:
    query = await session.execute(select(TriggerFormula.formula).where(
        TriggerFormula.id == pk,
    ))
    result = query.scalar()

    return result


async def get_actual_formulas():
    async with postgres_db.session_factory() as session:
        query = await session.execute(select(TriggerFormula.id, TriggerFormula.formula).where(
            TriggerFormula.is_active == True,
            TriggerFormula.is_shutted_off == False,
        ))
        result = query.fetchall()

    return result
