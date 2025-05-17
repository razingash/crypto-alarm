from fastapi import APIRouter, Depends, status
from sqlalchemy.ext.asyncio import AsyncSession
from apps.analytics.crud import get_formula_by_id, get_api_cooldown
from core.analysis.manager import formula_manager
from core.models import TriggerFormula
from db.postgre import postgres_db

router = APIRouter()

"""
message xchange with crypto-gateway service. implemented through Reast and not GRPC,
since in the future the project will be converted into SaaS
"""

@router.post(path='/formula/{pk}/')
async def add_formula(pk: int, session: AsyncSession = Depends(postgres_db.session_dependency)):
    """adds the formula to the graph"""
    formula = await get_formula_by_id(session, pk, TriggerFormula.formula)
    res = await formula_manager.add_formulas_to_graph(formula, pk)
    if not res:
        return {"error": res}
    return status.HTTP_200_OK


@router.delete(path='/formula/{pk}/')
async def remove_formula(pk: int):
    """удаляет формулу из графа"""
    res = await formula_manager.remove_formulas_from_graph(pk)
    if res is not True:
        return {"error": res}
    return status.HTTP_200_OK

@router.put(path='/formula/{pk}/')
async def update_formula(pk: int, session: AsyncSession = Depends(postgres_db.session_dependency)):
    """обновляет(удаляет и заново создает) формулу в графе"""
    formula_data = await get_formula_by_id(session, pk, TriggerFormula.formula, TriggerFormula.is_active)
    res = await formula_manager.update_formula_in_graph(pk, formula_data)

    if not res:
        return {"error": res}
    return status.HTTP_200_OK

@router.put(path='/endpoint/{pk}/')
async def update_api_cooldown(pk: int, session: AsyncSession = Depends(postgres_db.session_dependency)):
    """Changes the proxying frequency of a specific Binance API"""
    cooldown, api = await get_api_cooldown(session, pk)
    res = await formula_manager.update_api_frequency_cooldown(cooldown, api)

    if res is None:
        return status.HTTP_200_OK
    return {"error": res}
