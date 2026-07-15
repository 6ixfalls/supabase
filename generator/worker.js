const railwayApi = "https://backboard.railway.com/graphql/v2";

const cors = {
    "Access-Control-Allow-Origin": "https://6ixfalls.github.io",
    "Access-Control-Allow-Headers": "Authorization, Content-Type",
    "Access-Control-Allow-Methods": "POST, OPTIONS",
    "Access-Control-Max-Age": "86400",
};

function error(message, status) {
    return Response.json({ error: message }, { status, headers: cors });
}

export default {
    async fetch(request) {
        const url = new URL(request.url);
        if (url.pathname !== "/railway/graphql")
            return error("Not found.", 404);
        if (request.method === "OPTIONS")
            return new Response(null, { status: 204, headers: cors });
        if (request.method !== "POST")
            return error("Method not allowed.", 405);

        const authorization = request.headers.get("Authorization");
        if (!authorization?.startsWith("Bearer "))
            return error("Missing OAuth token.", 401);
        try {
            const body = await request.text();
            if (body.length > 64_000)
                return error("Request too large.", 413);
            const railway = await fetch(railwayApi, {
                method: "POST",
                headers: {
                    Authorization: authorization,
                    "Content-Type": "application/json",
                },
                body,
            });
            return new Response(railway.body, {
                status: railway.status,
                headers: {
                    ...cors,
                    "Content-Type": "application/json; charset=utf-8",
                },
            });
        } catch {
            return error("Railway API unavailable.", 502);
        }
    },
};
