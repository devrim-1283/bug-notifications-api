import { createContext, useContext } from 'react';
import type { Language } from './types';

export interface Translations {
  pageTitle: string;
  pageSubtitle: string;
  labelSite: string;
  labelTitle: string;
  labelCategory: string;
  labelDesc: string;
  labelPageUrl: string;
  contactToggle: string;
  labelFullName: string;
  labelPhone: string;
  labelEmail: string;
  labelImages: string;
  dropText: string;
  submitBtn: string;
  sending: string;
  errorGeneric: string;
  autoDetected: string;
  selectPlaceholder: string;
  catDesign: string;
  catFunctionality: string;
  catPerformance: string;
  catContent: string;
  catMobile: string;
  catSecurity: string;
  catOther: string;
  titlePlaceholderBug: string;
  titlePlaceholderRequest: string;
  descPlaceholderBug: string;
  descPlaceholderRequest: string;
  errSiteRequired: string;
  errTitleRequired: string;
  errCategoryRequired: string;
  errDescRequired: string;
  maxImages: string;
  siteSelectPlaceholder: string;
  typeBug: string;
  typeRequest: string;
  successTitle: string;
  successText: string;
  newReport: string;
  footerText: string;
}

const translations: Record<Language, Translations> = {
  tr: {
    pageTitle: 'DevrimSoft Geri Bildirim',
    pageSubtitle: 'Geri bildiriminizi paylaşın',
    labelSite: 'Site',
    labelTitle: 'Başlık',
    labelCategory: 'Kategori',
    labelDesc: 'Açıklama',
    labelPageUrl: "Sayfa URL'i",
    contactToggle: 'Sizinle iletişime geçelim mi?',
    labelFullName: 'Ad Soyad',
    labelPhone: 'Telefon',
    labelEmail: 'E-posta',
    labelImages: 'Görseller (Maks. 5)',
    dropText: 'Dosyaları sürükleyin veya tıklayın',
    submitBtn: 'Gönder',
    sending: 'Gönderiliyor...',
    errorGeneric: 'Bir hata oluştu. Lütfen tekrar deneyin.',
    autoDetected: 'Otomatik algılandı',
    selectPlaceholder: 'Seçin...',
    catDesign: 'Tasarım',
    catFunctionality: 'İşlevsellik',
    catPerformance: 'Performans',
    catContent: 'İçerik',
    catMobile: 'Mobil',
    catSecurity: 'Güvenlik',
    catOther: 'Diğer',
    titlePlaceholderBug: 'Hatanın kısa başlığını yazın',
    titlePlaceholderRequest: 'Önerinizin kısa başlığını yazın',
    descPlaceholderBug:
      'Hatayı detaylı olarak açıklayın. Ne yaptığınızda, ne olmasını beklediğinizde ve ne olduğunu belirtin.',
    descPlaceholderRequest:
      'Önerinizi detaylı olarak açıklayın. Ne istediğinizi ve neden faydalı olacağını belirtin.',
    errSiteRequired: 'Lütfen bir site seçiniz',
    errTitleRequired: 'Lütfen başlık giriniz',
    errCategoryRequired: 'Lütfen kategori seçiniz',
    errDescRequired: 'Lütfen açıklama giriniz',
    maxImages: 'En fazla 5 görsel yüklenebilir',
    siteSelectPlaceholder: 'Site seçin...',
    typeBug: 'Hata Bildirimi',
    typeRequest: 'Öneriler',
    successTitle: 'Teşekkürler!',
    successText:
      'Geri bildiriminiz başarıyla gönderildi. En kısa sürede değerlendirilecektir.',
    newReport: 'Yeni Bildirim',
    footerText: 'Powered by',
  },
  en: {
    pageTitle: 'DevrimSoft Feedback',
    pageSubtitle: 'Share your feedback',
    labelSite: 'Site',
    labelTitle: 'Title',
    labelCategory: 'Category',
    labelDesc: 'Description',
    labelPageUrl: 'Page URL',
    contactToggle: 'Would you like us to contact you?',
    labelFullName: 'Full Name',
    labelPhone: 'Phone',
    labelEmail: 'Email',
    labelImages: 'Screenshots (Max 5)',
    dropText: 'Drag files here or click to browse',
    submitBtn: 'Submit',
    sending: 'Sending...',
    errorGeneric: 'An error occurred. Please try again.',
    autoDetected: 'Auto-detected',
    selectPlaceholder: 'Select...',
    catDesign: 'Design',
    catFunctionality: 'Functionality',
    catPerformance: 'Performance',
    catContent: 'Content',
    catMobile: 'Mobile',
    catSecurity: 'Security',
    catOther: 'Other',
    titlePlaceholderBug: 'Short title for the bug',
    titlePlaceholderRequest: 'Short title for your suggestion',
    descPlaceholderBug:
      'Describe the bug in detail. What did you do, what did you expect, and what happened?',
    descPlaceholderRequest:
      'Describe your suggestion in detail. What would you like and why would it be useful?',
    errSiteRequired: 'Please select a site',
    errTitleRequired: 'Please enter a title',
    errCategoryRequired: 'Please select a category',
    errDescRequired: 'Please enter a description',
    maxImages: 'Maximum 5 images allowed',
    siteSelectPlaceholder: 'Select a site...',
    typeBug: 'Bug Report',
    typeRequest: 'Suggestions',
    successTitle: 'Thank you!',
    successText:
      'Your feedback has been submitted successfully. It will be reviewed shortly.',
    newReport: 'New Report',
    footerText: 'Powered by',
  },
  de: {
    pageTitle: 'DevrimSoft Feedback',
    pageSubtitle: 'Teilen Sie Ihr Feedback',
    labelSite: 'Webseite',
    labelTitle: 'Titel',
    labelCategory: 'Kategorie',
    labelDesc: 'Beschreibung',
    labelPageUrl: 'Seiten-URL',
    contactToggle: 'Sollen wir Sie kontaktieren?',
    labelFullName: 'Vollständiger Name',
    labelPhone: 'Telefon',
    labelEmail: 'E-Mail',
    labelImages: 'Bilder (Max. 5)',
    dropText: 'Dateien hierher ziehen oder klicken',
    submitBtn: 'Absenden',
    sending: 'Wird gesendet...',
    errorGeneric: 'Ein Fehler ist aufgetreten. Bitte versuchen Sie es erneut.',
    autoDetected: 'Automatisch erkannt',
    selectPlaceholder: 'Auswählen...',
    catDesign: 'Design',
    catFunctionality: 'Funktionalität',
    catPerformance: 'Leistung',
    catContent: 'Inhalt',
    catMobile: 'Mobil',
    catSecurity: 'Sicherheit',
    catOther: 'Sonstiges',
    titlePlaceholderBug: 'Kurzer Titel für den Fehler',
    titlePlaceholderRequest: 'Kurzer Titel für Ihren Vorschlag',
    descPlaceholderBug:
      'Beschreiben Sie den Fehler im Detail. Was haben Sie getan, was erwartet und was ist passiert?',
    descPlaceholderRequest:
      'Beschreiben Sie Ihren Vorschlag im Detail. Was wünschen Sie sich und warum wäre es nützlich?',
    errSiteRequired: 'Bitte wählen Sie eine Webseite',
    errTitleRequired: 'Bitte geben Sie einen Titel ein',
    errCategoryRequired: 'Bitte wählen Sie eine Kategorie',
    errDescRequired: 'Bitte geben Sie eine Beschreibung ein',
    maxImages: 'Maximal 5 Bilder erlaubt',
    siteSelectPlaceholder: 'Webseite auswählen...',
    typeBug: 'Fehlerbericht',
    typeRequest: 'Vorschläge',
    successTitle: 'Vielen Dank!',
    successText:
      'Ihr Feedback wurde erfolgreich gesendet. Es wird in Kürze bearbeitet.',
    newReport: 'Neuer Bericht',
    footerText: 'Powered by',
  },
  ru: {
    pageTitle: 'DevrimSoft Отзывы',
    pageSubtitle: 'Поделитесь отзывом',
    labelSite: 'Сайт',
    labelTitle: 'Заголовок',
    labelCategory: 'Категория',
    labelDesc: 'Описание',
    labelPageUrl: 'URL страницы',
    contactToggle: 'Хотите, чтобы мы связались с вами?',
    labelFullName: 'Полное имя',
    labelPhone: 'Телефон',
    labelEmail: 'Эл. почта',
    labelImages: 'Изображения (Макс. 5)',
    dropText: 'Перетащите файлы или нажмите',
    submitBtn: 'Отправить',
    sending: 'Отправка...',
    errorGeneric: 'Произошла ошибка. Попробуйте снова.',
    autoDetected: 'Определено автоматически',
    selectPlaceholder: 'Выбрать...',
    catDesign: 'Дизайн',
    catFunctionality: 'Функциональность',
    catPerformance: 'Производительность',
    catContent: 'Контент',
    catMobile: 'Мобильный',
    catSecurity: 'Безопасность',
    catOther: 'Другое',
    titlePlaceholderBug: 'Короткий заголовок ошибки',
    titlePlaceholderRequest: 'Короткий заголовок предложения',
    descPlaceholderBug:
      'Подробно опишите ошибку. Что вы делали, что ожидали и что произошло?',
    descPlaceholderRequest:
      'Подробно опишите предложение. Что вы хотите и почему это будет полезно?',
    errSiteRequired: 'Пожалуйста, выберите сайт',
    errTitleRequired: 'Пожалуйста, введите заголовок',
    errCategoryRequired: 'Пожалуйста, выберите категорию',
    errDescRequired: 'Пожалуйста, введите описание',
    maxImages: 'Максимум 5 изображений',
    siteSelectPlaceholder: 'Выбрать сайт...',
    typeBug: 'Ошибка',
    typeRequest: 'Предложения',
    successTitle: 'Спасибо!',
    successText:
      'Ваш отзыв успешно отправлен. Он будет рассмотрен в ближайшее время.',
    newReport: 'Новый отчёт',
    footerText: 'Powered by',
  },
  uk: {
    pageTitle: 'DevrimSoft Відгуки',
    pageSubtitle: 'Поділіться відгуком',
    labelSite: 'Сайт',
    labelTitle: 'Заголовок',
    labelCategory: 'Категорія',
    labelDesc: 'Опис',
    labelPageUrl: 'URL сторінки',
    contactToggle: "Бажаєте, щоб ми зв'язалися з вами?",
    labelFullName: "Повне ім'я",
    labelPhone: 'Телефон',
    labelEmail: 'Ел. пошта',
    labelImages: 'Зображення (Макс. 5)',
    dropText: 'Перетягніть файли або натисніть',
    submitBtn: 'Відправити',
    sending: 'Відправка...',
    errorGeneric: 'Сталася помилка. Будь ласка, спробуйте ще раз.',
    autoDetected: 'Визначено автоматично',
    selectPlaceholder: 'Обрати...',
    catDesign: 'Дизайн',
    catFunctionality: 'Функціональність',
    catPerformance: 'Продуктивність',
    catContent: 'Контент',
    catMobile: 'Мобільний',
    catSecurity: 'Безпека',
    catOther: 'Інше',
    titlePlaceholderBug: 'Короткий заголовок помилки',
    titlePlaceholderRequest: 'Короткий заголовок пропозиції',
    descPlaceholderBug:
      'Детально опишіть помилку. Що ви робили, що очікували і що сталося?',
    descPlaceholderRequest:
      'Детально опишіть пропозицію. Що ви бажаєте і чому це буде корисно?',
    errSiteRequired: 'Будь ласка, оберіть сайт',
    errTitleRequired: 'Будь ласка, введіть заголовок',
    errCategoryRequired: 'Будь ласка, оберіть категорію',
    errDescRequired: 'Будь ласка, введіть опис',
    maxImages: 'Максимум 5 зображень',
    siteSelectPlaceholder: 'Обрати сайт...',
    typeBug: 'Помилка',
    typeRequest: 'Пропозиції',
    successTitle: 'Дякуємо!',
    successText:
      'Ваш відгук успішно відправлено. Він буде розглянутий найближчим часом.',
    newReport: 'Новий звіт',
    footerText: 'Powered by',
  },
  es: {
    pageTitle: 'DevrimSoft Comentarios',
    pageSubtitle: 'Comparta sus comentarios',
    labelSite: 'Sitio',
    labelTitle: 'Título',
    labelCategory: 'Categoría',
    labelDesc: 'Descripción',
    labelPageUrl: 'URL de la página',
    contactToggle: '¿Desea que le contactemos?',
    labelFullName: 'Nombre completo',
    labelPhone: 'Teléfono',
    labelEmail: 'Correo electrónico',
    labelImages: 'Imágenes (Máx. 5)',
    dropText: 'Arrastre archivos aquí o haga clic',
    submitBtn: 'Enviar',
    sending: 'Enviando...',
    errorGeneric: 'Ocurrió un error. Por favor, inténtelo de nuevo.',
    autoDetected: 'Detectado automáticamente',
    selectPlaceholder: 'Seleccionar...',
    catDesign: 'Diseño',
    catFunctionality: 'Funcionalidad',
    catPerformance: 'Rendimiento',
    catContent: 'Contenido',
    catMobile: 'Móvil',
    catSecurity: 'Seguridad',
    catOther: 'Otro',
    titlePlaceholderBug: 'Título breve del error',
    titlePlaceholderRequest: 'Título breve de su sugerencia',
    descPlaceholderBug:
      'Describa el error en detalle. ¿Qué hizo, qué esperaba y qué ocurrió?',
    descPlaceholderRequest:
      'Describa su sugerencia en detalle. ¿Qué le gustaría y por qué sería útil?',
    errSiteRequired: 'Por favor, seleccione un sitio',
    errTitleRequired: 'Por favor, ingrese un título',
    errCategoryRequired: 'Por favor, seleccione una categoría',
    errDescRequired: 'Por favor, ingrese una descripción',
    maxImages: 'Máximo 5 imágenes permitidas',
    siteSelectPlaceholder: 'Seleccionar sitio...',
    typeBug: 'Informe de error',
    typeRequest: 'Sugerencias',
    successTitle: '¡Gracias!',
    successText:
      'Sus comentarios se han enviado correctamente. Se revisarán en breve.',
    newReport: 'Nuevo informe',
    footerText: 'Powered by',
  },
};

export const LANGUAGE_NAMES: Record<Language, string> = {
  tr: 'Türkçe',
  en: 'English',
  de: 'Deutsch',
  ru: 'Русский',
  uk: 'Українська',
  es: 'Español',
};

export const LANGUAGES: Language[] = ['tr', 'en', 'de', 'ru', 'uk', 'es'];

export function detectLanguage(): Language {
  const nav = (navigator.language || 'tr').toLowerCase();
  if (nav.startsWith('en')) return 'en';
  if (nav.startsWith('de')) return 'de';
  if (nav.startsWith('ru')) return 'ru';
  if (nav.startsWith('uk')) return 'uk';
  if (nav.startsWith('es')) return 'es';
  return 'tr';
}

export function t(lang: Language): Translations {
  return translations[lang];
}

interface I18nContextValue {
  lang: Language;
  setLang: (lang: Language) => void;
  t: Translations;
}

export const I18nContext = createContext<I18nContextValue>({
  lang: 'tr',
  setLang: () => {},
  t: translations.tr,
});

export function useI18n() {
  return useContext(I18nContext);
}
