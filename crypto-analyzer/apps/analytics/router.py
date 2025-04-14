from fastapi import APIRouter, Depends, status
from sqlalchemy.ext.asyncio import AsyncSession
from apps.analytics.crud import get_formula_by_id
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
    res = await formula_manager.update_formula_in_graph(formula_data, pk)

    if not res:
        return {"error": res}
    return status.HTTP_200_OK
