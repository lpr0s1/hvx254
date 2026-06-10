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
      const { message } = req.body;

      if (!message || typeof message !== "string") {
        return res.status(400).json({ error: "" });
      }

      lastMessage = message;

      // 🔥 Appel du webhook IFTTT
      await fetch(
        "https://maker.ifttt.com/trigger/api_message/with/key/tVEGob3-rpwoFAzaJA4gW",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ value1: message })
        }
      );

      return res.status(200).json({
        status: "OK",
        receivedMessage: lastMessage
      });

    } catch (e) {
      return res.status(500).json({ error: "" });
    }
  }

  return res.status(405).json({ error: "" });
}
