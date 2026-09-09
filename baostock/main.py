from http import HTTPStatus
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any
from urllib.parse import parse_qs, urlparse

from baostock_api import (
    __version__,
    health_endpoint,
    query_deposit_rate_data_endpoint,
    query_dupont_data_endpoint,
    query_adjust_factor_endpoint,
    query_balance_data_endpoint,
    query_cash_flow_data_endpoint,
    query_dividend_data_endpoint,
    query_all_stock_endpoint,
    query_forecast_report_endpoint,
    query_growth_data_endpoint,
    query_hs300_stocks_endpoint,
    query_loan_rate_data_endpoint,
    query_money_supply_data_month_endpoint,
    query_money_supply_data_year_endpoint,
    query_operation_data_endpoint,
    query_performance_express_report_endpoint,
    query_profit_data_endpoint,
    query_required_reserve_ratio_data_endpoint,
    query_history_k_data_plus_endpoint,
    query_history_index_k_data_plus_endpoint,
    query_stock_basic_endpoint,
    query_stock_industry_endpoint,
    query_sz50_stocks_endpoint,
    query_trade_dates_endpoint,
    usage_endpoint,
    query_zz500_stocks_endpoint,
)
from baostock_api.shared import (
    DAILY_LIMIT,
    HOST,
    PORT,
    json_response,
    make_error_payload,
    usage_counter,
)


class BaostockApiHandler(BaseHTTPRequestHandler):
    server_version = f"BaostockApi/{__version__}"

    def do_GET(self) -> None:
        parsed = urlparse(self.path)

        if parsed.path == "/health":
            health_endpoint.handle(self)
            return

        if parsed.path == "/usage":
            usage_endpoint.handle(self)
            return

        if parsed.path not in {
            "/query_adjust_factor",
            "/query_balance_data",
            "/query_cash_flow_data",
            "/query_deposit_rate_data",
            "/query_all_stock",
            "/query_dividend_data",
            "/query_dupont_data",
            "/query_forecast_report",
            "/query_growth_data",
            "/query_hs300_stocks",
            "/query_loan_rate_data",
            "/query_money_supply_data_month",
            "/query_money_supply_data_year",
            "/query_operation_data",
            "/query_performance_express_report",
            "/query_profit_data",
            "/query_required_reserve_ratio_data",
            "/query_history_k_data_plus",
            "/query_history_index_k_data_plus",
            "/query_stock_basic",
            "/query_stock_industry",
            "/query_sz50_stocks",
            "/query_trade_dates",
            "/query_zz500_stocks",
        }:
            json_response(
                self,
                HTTPStatus.NOT_FOUND,
                make_error_payload("route not found", "not_found", usage_counter.get_stats()),
            )
            return

        query = parse_qs(parsed.query)

        try:
            usage = usage_counter.consume()
        except RuntimeError:
            json_response(
                self,
                HTTPStatus.TOO_MANY_REQUESTS,
                make_error_payload(
                    f"daily limit exceeded: {DAILY_LIMIT}",
                    "daily_limit_exceeded",
                    usage_counter.get_stats(),
                ),
            )
            return

        try:
            if parsed.path == "/query_all_stock":
                payload = query_all_stock_endpoint.execute(
                    query_all_stock_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_adjust_factor":
                payload = query_adjust_factor_endpoint.execute(
                    query_adjust_factor_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_dividend_data":
                payload = query_dividend_data_endpoint.execute(
                    query_dividend_data_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_deposit_rate_data":
                payload = query_deposit_rate_data_endpoint.execute(
                    query_deposit_rate_data_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_dupont_data":
                payload = query_dupont_data_endpoint.execute(
                    query_dupont_data_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_forecast_report":
                payload = query_forecast_report_endpoint.execute(
                    query_forecast_report_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_operation_data":
                payload = query_operation_data_endpoint.execute(
                    query_operation_data_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_growth_data":
                payload = query_growth_data_endpoint.execute(
                    query_growth_data_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_hs300_stocks":
                payload = query_hs300_stocks_endpoint.execute(
                    query_hs300_stocks_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_loan_rate_data":
                payload = query_loan_rate_data_endpoint.execute(
                    query_loan_rate_data_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_balance_data":
                payload = query_balance_data_endpoint.execute(
                    query_balance_data_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_money_supply_data_month":
                payload = query_money_supply_data_month_endpoint.execute(
                    query_money_supply_data_month_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_money_supply_data_year":
                payload = query_money_supply_data_year_endpoint.execute(
                    query_money_supply_data_year_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_cash_flow_data":
                payload = query_cash_flow_data_endpoint.execute(
                    query_cash_flow_data_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_profit_data":
                payload = query_profit_data_endpoint.execute(
                    query_profit_data_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_performance_express_report":
                payload = query_performance_express_report_endpoint.execute(
                    query_performance_express_report_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_history_index_k_data_plus":
                payload = query_history_index_k_data_plus_endpoint.execute(
                    query_history_index_k_data_plus_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_required_reserve_ratio_data":
                payload = query_required_reserve_ratio_data_endpoint.execute(
                    query_required_reserve_ratio_data_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_stock_basic":
                payload = query_stock_basic_endpoint.execute(
                    query_stock_basic_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_stock_industry":
                payload = query_stock_industry_endpoint.execute(
                    query_stock_industry_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_sz50_stocks":
                payload = query_sz50_stocks_endpoint.execute(
                    query_sz50_stocks_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_trade_dates":
                payload = query_trade_dates_endpoint.execute(
                    query_trade_dates_endpoint.parse_params(query)
                )
            elif parsed.path == "/query_zz500_stocks":
                payload = query_zz500_stocks_endpoint.execute(
                    query_zz500_stocks_endpoint.parse_params(query)
                )
            else:
                payload = query_history_k_data_plus_endpoint.execute(
                    query_history_k_data_plus_endpoint.parse_params(query)
                )
        except ValueError as error:
            json_response(
                self,
                HTTPStatus.BAD_REQUEST,
                make_error_payload(str(error), "invalid_request", usage),
            )
            return
        except Exception as error:
            json_response(
                self,
                HTTPStatus.BAD_GATEWAY,
                make_error_payload(str(error), "baostock_query_failed", usage),
            )
            return

        json_response(
            self,
            HTTPStatus.OK,
            {
                "ok": True,
                "data": payload,
                "usage": usage.to_dict(),
            },
        )

    def log_message(self, format: str, *args: Any) -> None:
        print(f"[baostock-api] {self.address_string()} - {format % args}")


def main() -> None:
    server = ThreadingHTTPServer((HOST, PORT), BaostockApiHandler)
    print(f"baostock api running on http://{HOST}:{PORT}")
    server.serve_forever()


if __name__ == "__main__":
    main()
