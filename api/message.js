export default function handler(req, res) {
  // Réponse quand on fait un GET (afficher un message)
  if (req.method === "GET") {
    return res.status(200).json({
      status: "OK",
      message: "Ton API fonctionne H."
    });
  }

  // Réponse quand on fait un POST (envoyer un message)
  if (req.method === "POST") {
    const { message } = req.body;

    if (!message) {
      return res.status(400).json({ error: "Aucun message reçu" });
    }

    console.log("Message reçu :", message);

    return res.status(200).json({
      status: "OK",
      received: message
    });
  }

  // Si ce n'est ni GET ni POST
  return res.status(405).json({ error: "Méthode non autorisée" });
}
