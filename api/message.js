let lastMessage = "Aucun message pour le moment";

export default function handler(req, res) {
  if (req.method === "GET") {
    return res.status(200).json({
      status: "Disponible",
      message: lastMessage
    });
  }

  if (req.method === "POST") {
    const { message } = req.body || {};

    if (!message) {
      return res.status(400).json({ error: "Aucun message..." });
    }

    lastMessage = message;

    return res.status(200).json({
      status: "Disponible",
      received: message
    });
  }

  return res.status(405).json({ error: "405" });
}
