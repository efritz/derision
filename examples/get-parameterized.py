#!/usr/bin/env python3

import derision
import requests


def main():
    with derision.Client() as d:
        d.register(
            path='/users/(\d+)',
            response_body='You requested user {{ index .PathGroups 1 }}.',
        )

        resp = requests.get('http://localhost:5000/users/12345')
        assert resp.status_code == 200
        assert resp.text == 'You requested user 12345.'
        assert d.get_requests()[0].path == '/users/12345'

    print('Done.')


if __name__ == '__main__':
    main()
