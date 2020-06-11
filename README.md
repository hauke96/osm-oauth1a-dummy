# OSM OAuth 1.0a dummy server
<span style="color:red;"><b>Warning:</b> For development-use only!</span>

This server offers a very basic, simple and non-standardized, probably unbelievable buggy and insecure dummy implementation of the OAuth 1.0a authentication procedure used by the OpenStreetMap servers.
Therefore I use this to **be independent of the OSM servers**.

This also implements a very basic dummy API for the user information (so the `/users?users=...`, `user/details` and `/changesets` endpoints).
None of these APIs returns real data, just dummy data which has at least a somehow correct syntax.