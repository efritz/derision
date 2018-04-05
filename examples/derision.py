import attrdict
import requests


class Client(object):
    def __init__(self, addr='http://localhost:5000'):
        self.addr = addr

    def __enter__(self):
        return self

    def __exit__(self, *args):
        self.clear()

    def register(
        self,
        method=None,
        path=None,
        request_headers=None,
        status_code=None,
        response_headers=None,
        response_body=None,
    ):
        request = {
            'method': method,
            'path': path,
            'headers': request_headers,
        }

        response = {
            'status_code': status_code,
            'headers': response_headers,
            'body': response_body,
        }

        data = {
            'request': remove_none(request),
            'response': remove_none(response),
        }

        url = '{}/_control/register'.format(self.addr)
        resp = requests.post(url, json=data)
        resp.raise_for_status()

    def clear(self, ):
        url = '{}/_control/clear'.format(self.addr)
        resp = requests.post(url)
        resp.raise_for_status()

    def get_requests(self, clear=True):
        url = '{}/_control/requests{}'.format(
            self.addr,
            '?clear=true' if clear else '',
        )

        resp = requests.get(url)
        resp.raise_for_status()
        return [attrdict.AttrDict(r) for r in resp.json()]


def remove_none(d):
    return {k: v for (k, v) in d.items() if v is not None}
