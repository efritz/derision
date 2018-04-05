#!/usr/bin/env python3

import json
import derision
import requests


def main():
    with derision.Client() as d:
        requests.post('http://localhost:5000/', json={
            'foo': 'bar',
            'baz': 'bonk',
        })

        assert json.loads(d.get_requests()[0].body) == {
            'foo': 'bar',
            'baz': 'bonk',
        }

    print('Done.')


if __name__ == '__main__':
    main()
