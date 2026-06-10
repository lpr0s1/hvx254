let lastMessage = "";
export const config = {
  api: { bodyParser: true }
};

export default async function handler(req, res) {
  if (req.method === "GET") {
    return res.status(200).json({
      status: "Disponible",
      message: lastMessage
    });
  }
  if (req.method === "POST") {
    try {
      if (!req.body || typeof req.body.message !== "string") {
        return res.status(400).json({ error: "Invalid body" });
      }
      const message = req.body.message.trim();
      lastMessage = message;
      const iftttResponse = await fetch(
        "https://maker.ifttt.com/trigger/api_message/with/key/tVEGob3-rpwoFAzaJA4gW",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ value1: message })
        }
      );

      if (!iftttResponse.ok) {
        return res.status(500).json({ error: "IFTTT error" });
      }

      return res.status(200).json({
        status: "OK",
        receivedMessage: lastMessage
      });

    } catch (e) {
      return res.status(500).json({ error: "Server error" });
    }
  }

  return res.status(405).json({ error: "Method not allowed" });
}
