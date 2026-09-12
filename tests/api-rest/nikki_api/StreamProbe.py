"""Opens a streaming response and reports its headers without reading the body.

RequestsLibrary cannot do this. Every one of its keywords logs the response, and logging a
response reads `.text` — which on an endpoint whose body never ends blocks until the test
times out. `stream=True` does not help, because the read happens after the keyword returns.

So the notification stream is probed here instead: the request is made directly, the headers
are captured, and the connection is released without ever touching the body. What the suite
asserts against is therefore the status line and the headers, which is the whole of what a
single request can prove about an endpoint that is designed never to finish.
"""

import requests
import urllib3


class StreamProbe:
    ROBOT_LIBRARY_SCOPE = "GLOBAL"

    def open_stream(self, url, token=None, client_cert=None, client_key=None,
                    verify=True, timeout=10, **params):
        """Opens the stream and returns its status and headers.

        Returns a dict rather than the response object: the response is closed before this
        returns, so handing one back would invite a caller to read a body that is gone.
        """
        if verify in (False, "False", "false", "0", "no"):
            verify = False
            urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

        headers = {"Accept": "application/x-ndjson"}
        if token:
            headers["Authorization"] = "Bearer %s" % token

        cert = (client_cert, client_key) if client_cert and client_key else None
        query = {key: value for key, value in params.items() if value not in (None, "")}

        response = requests.get(
            url, headers=headers, params=query, cert=cert, verify=verify,
            stream=True, timeout=timeout,
        )
        try:
            return {
                "status_code": response.status_code,
                "headers": dict(response.headers),
            }
        finally:
            # Closing without reading is the point. The backend holds a response open for as
            # long as the client is there, so a probe that walked away would leave a stream
            # running on the server for every test in the suite.
            response.close()
