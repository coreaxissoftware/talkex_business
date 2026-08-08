"""Central exception -> JSON response mapping, so every router returns the
same error envelope shape instead of each one hand-rolling it."""
from fastapi import FastAPI, Request, status
from fastapi.responses import JSONResponse


def register_error_handlers(app: FastAPI) -> None:
    @app.exception_handler(Exception)
    async def unhandled_exception_handler(request: Request, exc: Exception):
        return JSONResponse(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            content={"detail": "Internal server error"},
        )
