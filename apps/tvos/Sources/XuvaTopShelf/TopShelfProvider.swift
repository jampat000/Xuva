import TVServices

@objc(TopShelfProvider)
class TopShelfProvider: NSObject, TVTopShelfContentProvider {
    func topShelfItems(completionHandler: @escaping (TVTopShelfContent?) -> Void) {
        guard let defaults = UserDefaults(suiteName: topShelfAppGroup),
              let data = defaults.data(forKey: topShelfPayloadKey),
              let payload = try? JSONDecoder().decode(TopShelfPayload.self, from: data),
              !payload.items.isEmpty else {
            completionHandler(nil)
            return
        }

        let carouselItems = payload.items.map { entry -> TVTopShelfCarouselItem in
            let item = TVTopShelfCarouselItem(identifier: entry.id)
            item.title = entry.title
            if let url = entry.localImageURL, FileManager.default.fileExists(atPath: url.path) {
                item.setImageURL(url, for: .screenScale1x)
                item.setImageURL(url, for: .screenScale2x)
            }
            if let url = entry.deepLinkURL {
                item.displayAction = TVTopShelfAction(url: url)
                item.playAction = TVTopShelfAction(url: url)
            }
            return item
        }

        completionHandler(TVTopShelfCarouselContent(style: .actions, items: carouselItems))
    }
}
